package utils_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-project/modulectl/internal/utils"
)

func Test_GetImageNameAndTag(t *testing.T) {
	tests := []struct {
		name              string
		imageURL          string
		expectedImageName string
		expectedImageTag  string
		expectedError     error
	}{
		{
			name:              "valid image URL",
			imageURL:          "docker.io/template-operator/test:latest",
			expectedImageName: "test",
			expectedImageTag:  "latest",
			expectedError:     nil,
		},
		{
			name:              "invalid image URL - no tag",
			imageURL:          "docker.io/template-operator/test",
			expectedImageName: "",
			expectedImageTag:  "",
			expectedError:     errors.New("invalid image URL"),
		},
		{
			name:              "invalid image URL - multiple tags",
			imageURL:          "docker.io/template-operator/test:latest:latest",
			expectedImageName: "",
			expectedImageTag:  "",
			expectedError:     errors.New("invalid image URL"),
		},
		{
			name:              "invalid image URL - no slashes",
			imageURL:          "docker.io",
			expectedImageName: "",
			expectedImageTag:  "",
			expectedError:     errors.New("invalid image URL"),
		},
		{
			name:              "invalid image URL - empty URL",
			imageURL:          "",
			expectedImageName: "",
			expectedImageTag:  "",
			expectedError:     errors.New("invalid image URL"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			imgName, imgTag, err := utils.GetImageNameAndTag(test.imageURL)
			if test.expectedError != nil {
				require.ErrorContains(t, err, test.expectedError.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expectedImageName, imgName)
				require.Equal(t, test.expectedImageTag, imgTag)
			}
		})
	}
}
