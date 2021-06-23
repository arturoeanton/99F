package jsonschema

import (
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/couchbase/gocb/v2"
)

var (
	//Cluster ...
	Cluster *gocb.Cluster
)

// InitDB ..
func InitDB() {
	var err error
	Cluster, err = gocb.Connect(
		"localhost",
		gocb.ClusterOptions{
			Username: "Administrator",
			Password: "seaside!!",
		})
	if err != nil {
		log.Fatalln(err.Error())
		return
	}
}

// CreateBucket ...
func CreateBucket(name string) {
	_, err := Cluster.Buckets().GetBucket(name, nil)
	if err != nil {
		bucketSettings := gocb.CreateBucketSettings{
			BucketSettings: gocb.BucketSettings{
				Name:                 name,
				FlushEnabled:         false,
				ReplicaIndexDisabled: true,
				RAMQuotaMB:           200,
				NumReplicas:          1,
				BucketType:           gocb.CouchbaseBucketType,
			},
			ConflictResolutionType: gocb.ConflictResolutionTypeSequenceNumber,
		}
		err = Cluster.Buckets().CreateBucket(bucketSettings, nil)
		if err != nil {
			log.Fatalln(err.Error())
		}
		log.Println("Create Bucket -> " + name)
	}
	bucket := Cluster.Bucket(name)
	err = bucket.WaitUntilReady(20*time.Second, nil)
	if err != nil {
		log.Fatalln(err.Error())
	}
	log.Println("Ready Bucket -> " + name)
	Cluster.QueryIndexes().CreatePrimaryIndex(name, &gocb.CreatePrimaryQueryIndexOptions{IgnoreIfExists: true})
}

func CreateBuckets(folderConfig string) {
	files, err := ioutil.ReadDir(folderConfig)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".json") {
			name := f.Name()
			CreateBucket(name[0 : len(name)-5])
		}
	}
}
