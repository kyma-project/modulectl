package utils

import (
	"errors"
	"fmt"
	"strings"
)

var errInvalidURL = errors.New("invalid image URL")

const imageTagSlicesLength = 2

func GetImageNameAndTag(imageURL string) (string, string, error) {
	imageTag := strings.Split(imageURL, ":")
	if len(imageTag) != imageTagSlicesLength {
		return "", "", fmt.Errorf("image URL: %s: %w", imageURL, errInvalidURL)
	}

	imageName := strings.Split(imageTag[0], "/")
	if len(imageName) == 0 {
		return "", "", fmt.Errorf("image URL: %s: %w", imageURL, errInvalidURL)
	}

	return imageName[len(imageName)-1], imageTag[len(imageTag)-1], nil
}
