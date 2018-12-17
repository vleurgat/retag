package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os/user"
	"path"
	"strings"

	"github.com/vleurgat/dockerclient/pkg/dockerclient"
	"github.com/vleurgat/regstat/internal/app/docker"
)

func splitTag(tag string) (string, string, string, error) {
	split1 := strings.Split(tag, "/")
	if len(split1) != 2 {
		return "", "", "", errors.New("invalid tag " + tag)
	}
	split2 := strings.Split(split1[1], ":")
	if len(split2) != 2 {
		return "", "", "", errors.New("invalid tag " + tag)
	}
	return split1[0], split2[0], split2[1], nil
}

func registryURL(registry string, repo string, tag string) string {
	return fmt.Sprintf("http://%s/v2/%s/manifests/%s", registry, repo, tag)
}

func main() {
	var fromTag string
	var toTag string
	var dockerConfigFile string
	flag.StringVar(&fromTag, "from-tag", "", "the existing image tag")
	flag.StringVar(&toTag, "to-tag", "", "the new image tag - must have the same registry and repo")
	flag.StringVar(&dockerConfigFile, "docker-config", "", "the path to the Docker registry config.json file, used to obtain login credentials")
	flag.Parse()
	if fromTag == "" || toTag == "" {
		log.Fatal("Error: must supply both from-tag and to-tag")
	}
	registry1, repo1, tag1, err := splitTag(fromTag)
	if err != nil {
		log.Fatal("failed to parse from-tag", err)
	}
	registry2, repo2, tag2, err := splitTag(toTag)
	if err != nil {
		log.Fatal("failed to parse to-tag", err)
	}
	if registry1 != registry2 || repo1 != repo2 {
		log.Fatalf("registry and repo must be the same: %s != %s, or %s != %s", registry1, registry2, repo1, repo2)
	}
	if tag1 == tag2 {
		log.Fatalf("tags must be different: %s == %s", tag1, tag2)
	}
	if dockerConfigFile == "" {
		usr, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		dockerConfigFile = path.Join(usr.HomeDir, ".docker", "config.json")
	}
	dockerConfig, err := docker.CreateConfig(dockerConfigFile)

	log.Printf("from %s (%s, %s, %s)", fromTag, registry1, repo1, tag1)
	log.Printf("to %s (%s, %s, %s)", toTag, registry2, repo2, tag2)
	log.Println("using " + dockerConfigFile)

	url1 := registryURL(registry1, repo1, tag1)
	url2 := registryURL(registry2, repo2, tag2)
	log.Println("with from url", url1)
	log.Println("with to url", url2)

	client := dockerclient.CreateClient(dockerConfig)
	manifest, err := client.GetV2Manifest(url1)
	if err != nil {
		log.Fatalf("failed to get manifest for %s, %s", fromTag, err.Error())
	}
	err = client.PutV2Manifest(url2, manifest)
	if err != nil {
		log.Fatalf("failed to PUT manifest for %s, %s", toTag, err.Error())
	}
	log.Println("success")
}
