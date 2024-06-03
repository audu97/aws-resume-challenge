package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"io/fs"
	"mime"
	"path"
	"path/filepath"
	"reflect"
)

type Folder struct {
	pulumi.ResourceState

	bucketName pulumi.IDOutput     `pulumi:"bucketName"`
	websiteUrl pulumi.StringOutput `pulumi:"websiteUrl"`
}

func NewS3Folder(ctx *pulumi.Context, bucketName string, siteDir string, args *FolderArgs) (*Folder, error) {

	var resource Folder
	// Stack exports
	err := ctx.RegisterComponentResource("pulumi:example:S3Folder", bucketName, &resource)
	if err != nil {
		return nil, err
	}
	// Create a bucket and expose a website index document
	siteBucket, err := s3.NewBucket(ctx, bucketName, &s3.BucketArgs{
		Website: s3.BucketWebsiteArgs{
			IndexDocument: pulumi.String("index.html"),
		},
	}, pulumi.Parent(&resource))
	if err != nil {
		return nil, err
	}

	// For each file in the directory, create an S3 object stored in `siteBucket`
	err = filepath.Walk(siteDir, func(name string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			rel, err := filepath.Rel(siteDir, name)
			if err != nil {
				return err
			}

			if _, err := s3.NewBucketObject(ctx, rel, &s3.BucketObjectArgs{
				Bucket:      siteBucket.ID(),                                     // reference to the s3.Bucket object
				Source:      pulumi.NewFileAsset(name),                           // use FileAsset to point to a file
				ContentType: pulumi.String(mime.TypeByExtension(path.Ext(name))), // set the MIME type of the file
			}, pulumi.Parent(&resource)); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &resource, nil
}

type folderArgs struct {
}

type FolderArgs struct {
}

func (FolderArgs) ElementType() reflect.Type {

	return reflect.TypeOf((*folderArgs)(nil)).Elem()
}
