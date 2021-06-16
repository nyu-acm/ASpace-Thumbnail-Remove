package main

import (
	"flag"
	"github.com/nyudlts/go-aspace"
	"log"
	"os"
)

var (
	client *aspace.ASClient
  	err error
	repoId int
	test bool
)

func init() {
	flag.IntVar(&repoId, "repository", 0, "the repository")
	flag.BoolVar(&test, "test", false, "test mode")
}
func main() {
	f, err := os.OpenFile("thumbnail-removal.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)

	log.Println("INFO", "do-migration tool")
	flag.Parse()
	client, err = aspace.NewClient("/etc/go-aspace.yml", "fade", 20)
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
			for _, i := range thumbs {
				log.Printf("INFO fv #%d is a thumbnail", i)
			}
		}
	} else if IsDOThumbnailOnly(do.FileVersions) == true {
		log.Printf("INFO Deleting thumbnail-only digital object %s %s\n", do.URI, do.Title)
		if test != true {
			msg, err := client.DeleteDigitalObject(repoId, doId)
			if err != nil {
				log.Printf("ERROR Failed to delete digital object %s\n", do.URI)
				return nil
			}
			log.Printf("INFO %s deleted: %s\n", do.URI, msg)
		}
	} else {
		log.Printf("INFO %s conforms to all rules, skipping\n", do.URI)
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

