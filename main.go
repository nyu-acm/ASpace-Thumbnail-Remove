package main

import (
	"flag"
	"fmt"
	"github.com/nyudlts/go-aspace"
	"log"
	"os"
)

var (
	client *aspace.ASClient
  	err error
	repoId int
	environment string
	test bool
)

func init() {
	flag.IntVar(&repoId, "repository", 0, "the repository")
	flag.StringVar(&environment, "environment", "", "the environemnt")
	flag.BoolVar(&test, "test", false, "test mode")
}

func main() {
	flag.Parse()
	logfilename := fmt.Sprintf("thumbnail-removal-repository-%d.log", repoId)

	fmt.Println("Running, logging to", logfilename)
	f, err := os.OpenFile(logfilename, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error creating logfile: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)
	log.Println("INFO", "thumbnail remove tool")


	client, err = aspace.NewClient("/etc/go-aspace.yml", environment, 20)
	if err != nil {
		log.Println("ERROR", err)
		log.Fatal("FATAL", "Could not get a client")
	}
	log.Println("INFO", aspace.LibraryVersion)
	doIds, err := client.GetDigitalObjectIDs(repoId)
	if err != nil {
		log.Fatal("FATAL ", "Could not get a list of DO IDs for repository",  repoId)
	}

	for _, doId := range doIds {
		err := processDigitalObject(repoId, doId); if err != nil {
			log.Println("ERROR", err.Error())
		}
	}

}

func processDigitalObject(repoId int, doId int) error {
	do, err := client.GetDigitalObject(repoId, doId)
	if err != nil {
		log.Printf("ERROR Failed to get digital object %s\n", do.URI)
		return nil
	}
	log.Printf("INFO Processing %s", do.URI)
	//check if the do contains more than one file version if it does -- skip
	if len(do.FileVersions) > 1 {
		log.Printf("INFO %s contains more than one file version", do.URI)
		thumbs := containsThumbnail(do.FileVersions)
		if len(thumbs) > 0 {
			var newFV []aspace.FileVersion
			for _, i := range thumbs {
				log.Printf("INFO Removing thumbnail file version #%d", i)
				newFV = removeFileVersion(do.FileVersions, i)
			}

			do.FileVersions = newFV

			msg, err := client.UpdateDigitalObject(repoId, doId, do)
			if err != nil {
				log.Printf("ERROR Failed to update digital object %s", do.URI)
				return nil
			}
			log.Printf("INFO %s updated: %s", do.URI, msg)
		}

	} else if IsDOThumbnailOnly(do.FileVersions) == true {
		log.Printf("INFO Deleting thumbnail-only digital object %s %s\n", do.URI, do.Title)
		if test != true {
			msg, err := client.DeleteDigitalObject(repoId, doId)
			if err != nil {
				log.Printf("ERROR Failed to delete digital object %s\n", do.URI)
				return nil
			}
			log.Printf("INFO %s deleted: %s", do.URI, msg)
		}
	} else {
		log.Printf("INFO %s conforms to all rules, skipping", do.URI)
	}

	return nil
}

func containsThumbnail(fvs []aspace.FileVersion) []int {
	thumbnails := [] int{}
	for i, fv := range fvs {
		if fv.UseStatement == "image-thumbnail" {
			thumbnails = append(thumbnails, i)
		}
	}
	return thumbnails
}

func IsDOThumbnailOnly(fileversions []aspace.FileVersion) bool {
	if len(fileversions) == 1 {
		if fileversions[0].UseStatement == "image-thumbnail" {
			return true
		}
	}
	return false
}

func removeFileVersion(fvs []aspace.FileVersion, i int) []aspace.FileVersion {
	fvs[i] = fvs[len(fvs)-1]
	return fvs[:len(fvs)-1]
}

