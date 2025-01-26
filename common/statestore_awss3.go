package common

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"

	"github.com/apex/log"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type awsS3StateStoreCredentials struct {
	awsAccesskey    string // The AWS access key to use for the state store creation.
	awsSecretkey    string // The AWS secret key to use for the state store creation.
	awsSessionToken string // The AWS session token to use for the state store creation (optional).
}

type awsS3StateStore struct {
	baseName       string                     // The name of the bucket to use for the state store.
	bucketRegion   string                     // The AWS region to use for the state store creation.
	bucketTags     map[string]string          // The tags to apply to the bucket.
	awsCredentials awsS3StateStoreCredentials // The AWS credentials to use for the state store creation.
	awsAPIClient   *s3.Client                 // The AWS API client to use for the state store creation.
	awsAPIContext  context.Context            // The AWS API context to use for the state store creation.
}

// Create state store with the well-known credentials from the environment.
func GetAWSS3StateStore(baseName string, bucketTags map[string]string) (StateStore, error) {
	awsCredentials := awsS3StateStoreCredentials{
		awsAccesskey:    os.Getenv("AWS_ACCESS_KEY_ID"),
		awsSecretkey:    os.Getenv("AWS_SECRET_ACCESS_KEY"),
		awsSessionToken: os.Getenv("AWS_SESSION_TOKEN"),
	}

	bucketRegion := os.Getenv("AWS_DEFAULT_REGION")
	if bucketRegion == "" {
		bucketRegion = "eu-central-1"
	}

	awsAPIContext := context.TODO()

	awsAPIConfig, err := config.LoadDefaultConfig(awsAPIContext, config.WithRegion(bucketRegion))
	if err != nil {
		return nil, err
	}

	awsAPIClient := s3.NewFromConfig(awsAPIConfig)

	return &awsS3StateStore{
		baseName:       baseName,
		bucketRegion:   bucketRegion,
		bucketTags:     bucketTags,
		awsCredentials: awsCredentials,
		awsAPIClient:   awsAPIClient,
		awsAPIContext:  awsAPIContext,
	}, nil
}

func (ss *awsS3StateStore) BucketName() string {
	return ss.baseName + "-state"
}

// Creates and/or logs in to a state store and returns its URL string.
func (ss *awsS3StateStore) StoreOpen() (string, error) {
	// Does the bucket exist?
	exists, err := ss.bucketExists()
	if err != nil {
		return "", err
	}

	// If the bucket does not exist, create it
	if !exists {
		if err := ss.createBucket(); err != nil {
			return "", err
		}
	}

	// Bucket is accessible?
	err = ss.bucketIsAccessible()
	if err != nil {
		return "", err
	}

	// Bucket exists, or was created successfully, so log in to it
	stateURI := "s3://" + ss.BucketName()

	// TODO: get this from the config
	err = os.Setenv("PULUMI_CONFIG_PASSPHRASE", "we need a persistent token for the project here")
	if err != nil {
		return "", err
	}

	// validate stateUri
	_, err = url.ParseRequestURI(stateURI)
	if err != nil {
		panic(err)
	}

	_, err = exec.Command("pulumi", "login", stateURI).Output() // #nosec G204
	if err != nil {
		return "", err
	}

	return stateURI, nil
}

// Closes or logs out of a state store, without deleting any data.
func (ss *awsS3StateStore) StoreClose() error {
	// Bucket is accessible?
	err := ss.bucketIsAccessible()
	if err != nil {
		return err
	}

	stateURI := "s3://" + ss.BucketName()

	// validate stateUri
	_, err = url.ParseRequestURI(stateURI)
	if err != nil {
		panic(err)
	}

	_, err = exec.Command("pulumi", "logout", stateURI).Output() // #nosec G204
	if err != nil {
		return err
	}

	return nil
}

// Deletes the state store, including all data when the force parameter is true.
func (ss *awsS3StateStore) StoreDelete(force bool) error {
	err := ss.StoreClose()
	if err != nil {
		return err
	}

	err = ss.deleteBucket(force)
	if err != nil {
		return err
	}

	return nil
}

func (ss *awsS3StateStore) deleteBucket(force bool) error {
	log.WithFields(log.Fields{
		"bucket": ss.BucketName(),
		"region": ss.bucketRegion,
	}).Debug("AWSS3StateStore deleting bucket")

	argBucketName := ss.BucketName()

	if force {
		deleteObject := func(bucket, key, versionId *string) error {
			log.WithFields(log.Fields{
				"bucket":  ss.BucketName(),
				"region":  ss.bucketRegion,
				"key":     *key,
				"version": aws.ToString(versionId),
			}).Debug("AWSS3StateStore deleting object")

			_, err := ss.awsAPIClient.DeleteObject(ss.awsAPIContext, &s3.DeleteObjectInput{
				Bucket:    bucket,
				Key:       key,
				VersionId: versionId,
			})
			if err != nil {
				log.WithFields(log.Fields{
					"bucket":  ss.BucketName(),
					"region":  ss.bucketRegion,
					"key":     *key,
					"version": aws.ToString(versionId),
				}).WithError(err).Error("AWSS3StateStore failed to delete object")

				return fmt.Errorf("error deleting object %s/%s: %s", *key, aws.ToString(versionId), err)
			}

			return nil
		}

		argsListObjects := &s3.ListObjectsV2Input{Bucket: &argBucketName}

		for {
			out, err := ss.awsAPIClient.ListObjectsV2(ss.awsAPIContext, argsListObjects)
			if err != nil {
				log.WithFields(log.Fields{
					"bucket": ss.BucketName(),
					"region": ss.bucketRegion,
				}).WithError(err).Error("AWSS3StateStore failed to list objects")

				return err
			}

			for _, item := range out.Contents {
				if err := deleteObject(&argBucketName, item.Key, nil); err != nil {
					return err
				}
			}

			if out.IsTruncated {
				argsListObjects.ContinuationToken = out.ContinuationToken
			} else {
				break
			}
		}

		argsListObjectVersions := &s3.ListObjectVersionsInput{Bucket: &argBucketName}

		for {
			out, err := ss.awsAPIClient.ListObjectVersions(ss.awsAPIContext, argsListObjectVersions)
			if err != nil {
				log.WithFields(log.Fields{
					"bucket": ss.BucketName(),
					"region": ss.bucketRegion,
				}).WithError(err).Error("AWSS3StateStore failed to list object versions")

				return fmt.Errorf("listObjectVersions failed to list versions: %v", err)
			}

			for _, item := range out.DeleteMarkers {
				log.WithFields(log.Fields{
					"bucket": ss.BucketName(),
					"region": ss.bucketRegion,
				}).Debug("AWSS3StateStore deleting object markers")

				if err := deleteObject(&argBucketName, item.Key, item.VersionId); err != nil {
					return err
				}
			}

			for _, item := range out.Versions {
				log.WithFields(log.Fields{
					"bucket": ss.BucketName(),
					"region": ss.bucketRegion,
				}).Debug("AWSS3StateStore deleting object versions")

				if err := deleteObject(&argBucketName, item.Key, item.VersionId); err != nil {
					return err
				}
			}

			if out.IsTruncated {
				argsListObjectVersions.VersionIdMarker = out.NextVersionIdMarker
				argsListObjectVersions.KeyMarker = out.NextKeyMarker
			} else {
				break
			}
		}
	}

	// Delete the bucket
	deleteArgs := &s3.DeleteBucketInput{
		Bucket: &argBucketName,
	}

	if _, err := ss.awsAPIClient.DeleteBucket(ss.awsAPIContext, deleteArgs); err != nil {
		log.WithFields(log.Fields{
			"bucket": ss.BucketName(),
			"region": ss.bucketRegion,
		}).WithError(err).Error("AWSS3StateStore failed to delete bucket")

		return err
	}

	log.WithFields(log.Fields{
		"bucket": ss.BucketName(),
		"region": ss.bucketRegion,
	}).Debug("AWSS3StateStore bucket deleted")

	return nil
}

func (ss *awsS3StateStore) createBucket() error {
	log.WithFields(log.Fields{
		"bucket": ss.BucketName(),
		"region": ss.bucketRegion,
	}).Debug("AWSS3StateStore creating bucket")

	argBucketName := ss.BucketName()

	createArgs := &s3.CreateBucketInput{
		Bucket: &argBucketName,
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(ss.bucketRegion),
		},
	}

	bucket, err := ss.awsAPIClient.CreateBucket(ss.awsAPIContext, createArgs)
	if err != nil {
		log.WithFields(log.Fields{
			"bucket": ss.BucketName(),
			"region": ss.bucketRegion,
		}).WithError(err).Error("AWSS3StateStore failed to create bucket")

		return err
	}

	log.WithFields(log.Fields{
		"bucket": ss.BucketName(),
		"region": ss.bucketRegion,
	}).Debugf("AWSS3StateStore created bucket %s", *bucket.Location)

	// Tag the bucket
	if len(ss.bucketTags) > 0 {
		log.WithFields(log.Fields{
			"bucket": ss.BucketName(),
			"region": ss.bucketRegion,
		}).Debugf("AWSS3StateStore tagging bucket with %d tags", len(ss.bucketTags))

		tagSet := []types.Tag{}
		for k, v := range ss.bucketTags {
			tagSet = append(tagSet, types.Tag{
				Key:   aws.String(k),
				Value: aws.String(v),
			})

			log.WithFields(log.Fields{
				"bucket": ss.BucketName(),
				"region": ss.bucketRegion,
				"key":    k,
				"value":  v,
			}).Debug("AWSS3StateStore tagging bucket")
		}

		tagArgs := &s3.PutBucketTaggingInput{
			Bucket: &argBucketName,
			Tagging: &types.Tagging{
				TagSet: tagSet,
			},
		}

		if _, err := ss.awsAPIClient.PutBucketTagging(ss.awsAPIContext, tagArgs); err != nil {
			log.WithFields(log.Fields{
				"bucket": ss.BucketName(),
				"region": ss.bucketRegion,
			}).WithError(err).Error("AWSS3StateStore failed to tag bucket")

			return err
		}
	}

	// Apply bucket Encryption
	encryptionType := types.ServerSideEncryptionAes256

	encryptionArgs := &s3.PutBucketEncryptionInput{
		Bucket: &argBucketName,
		ServerSideEncryptionConfiguration: &types.ServerSideEncryptionConfiguration{
			Rules: []types.ServerSideEncryptionRule{
				{
					ApplyServerSideEncryptionByDefault: &types.ServerSideEncryptionByDefault{
						SSEAlgorithm: encryptionType,
					},
				},
			},
		},
	}

	log.WithFields(log.Fields{
		"bucket": ss.BucketName(),
		"region": ss.bucketRegion,
		"type":   encryptionType,
	}).Debug("AWSS3StateStore applying bucket encryption")

	if _, err := ss.awsAPIClient.PutBucketEncryption(ss.awsAPIContext, encryptionArgs); err != nil {
		log.WithFields(log.Fields{
			"bucket": ss.BucketName(),
			"region": ss.bucketRegion,
		}).WithError(err).Error("AWSS3StateStore failed to apply bucket encryption")

		return err
	}

	return nil
}

func (ss *awsS3StateStore) bucketIsAccessible() error {
	argBucketName := ss.BucketName()

	// Check if the bucket is accessible
	headArgs := &s3.HeadBucketInput{
		Bucket: &argBucketName,
	}

	if _, err := ss.awsAPIClient.HeadBucket(ss.awsAPIContext, headArgs); err != nil {
		log.WithFields(log.Fields{
			"bucket": ss.BucketName(),
			"region": ss.bucketRegion,
		}).WithError(err).Error("AWSS3StateStore bucket is not accessible")

		return err
	}

	log.WithFields(log.Fields{
		"bucket": ss.BucketName(),
		"region": ss.bucketRegion,
	}).Debug("AWSS3StateStore bucket accessible")

	return nil
}

func (ss *awsS3StateStore) bucketExists() (bool, error) {
	bucketList, err := ss.awsAPIClient.ListBuckets(ss.awsAPIContext, &s3.ListBucketsInput{})
	if err != nil {
		log.WithFields(log.Fields{
			"bucket": ss.BucketName(),
			"region": ss.bucketRegion,
		}).WithError(err).Error("AWSS3StateStore failed to list buckets")

		return false, err
	}

	// Get a list of all the buckets in the account
	for _, bucket := range bucketList.Buckets {
		// Is the bucket name the same as the one we want?
		if *bucket.Name == ss.BucketName() {
			log.WithFields(log.Fields{
				"bucket": ss.BucketName(),
				"region": ss.bucketRegion,
			}).Debugf("AWSS3StateStore bucket exists")

			return true, nil
		}
	}

	// Bucket with given name does not exist
	log.WithFields(log.Fields{
		"bucket": ss.BucketName(),
		"region": ss.bucketRegion,
	}).Debugf("AWSS3StateStore bucket does not exist")

	return false, nil
}
