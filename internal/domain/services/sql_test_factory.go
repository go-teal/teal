package services

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	internalmodels "github.com/go-teal/teal/internal/domain/internal_models"
	"github.com/go-teal/teal/internal/domain/utils"
	"github.com/go-teal/teal/pkg/configs"
	"gopkg.in/yaml.v2"
)

const TESTS_DIR = "assets/tests"

func InitSQLTestsConfigs(config *configs.Config, projectProfile *configs.ProjectProfile) ([]*internalmodels.TestConfig, error) {
	var testConfigs []*internalmodels.TestConfig
	testProjectDir := config.ProjectPath + "/" + TESTS_DIR
	modelsProjectDir := config.ProjectPath + "/" + MODEL_DIR
	testFileNames, err := os.ReadDir(testProjectDir)
	if err != nil {
		fmt.Printf("Error reading directory: %v", err)
		return nil, err
	}

	for _, fileEntry := range testFileNames {
		if !fileEntry.IsDir() {
			testConfigs = append(testConfigs, initTestConfig("root",
				testProjectDir+"/"+fileEntry.Name(),
				fileEntry.Name(),
				projectProfile,
				modelsProjectDir))
		}
	}

	for _, stage := range projectProfile.Models.Stages {
		stageName := stage.Name
		// Read the directory contents
		stageTestFileNames, err := os.ReadDir(testProjectDir + "/" + stageName)
		if err != nil {
			continue
		}
		for _, stageFile := range stageTestFileNames {
			if !stageFile.IsDir() {
				testConfigs = append(testConfigs, initTestConfig(
					stageName,
					testProjectDir+"/"+stageName+"/"+stageFile.Name(),
					stageFile.Name(),
					projectProfile,
					modelsProjectDir))
			}
		}
	}

	return testConfigs, nil
}

func initTestConfig(
	stage string,
	fullPath string,
	fileName string,
	projectProfile *configs.ProjectProfile,
	modelsProjectDir string,
) *internalmodels.TestConfig {
	testFileByte, err := os.ReadFile(fullPath)
	goFuncName, refName := utils.CreateModelName(stage, fileName)
	if err != nil {
		panic(err)
	}
	testFileFinalTemplate, _, err := prepareModelTemplate(testFileByte, refName, modelsProjectDir, projectProfile)
	if err != nil {
		fmt.Printf("can not parse test profile %s\n", string(testFileByte))
		panic(err)
	}

	var inlineTestProfileByteBuffer bytes.Buffer
	globalTestProfile := projectProfile.GetTestProfile(stage, fileName)
	err = testFileFinalTemplate.ExecuteTemplate(&inlineTestProfileByteBuffer, "profile.yaml", nil)
	if err == nil {
		var newTestPrifile configs.TestProfile
		err = yaml.Unmarshal(inlineTestProfileByteBuffer.Bytes(), &newTestPrifile)
		if err != nil {
			fmt.Printf("can not unmarshal test profile")
			panic(err)
		}
		globalTestProfile.Connection = newTestPrifile.Connection
	}

	var sqlByteBuffer bytes.Buffer
	err = testFileFinalTemplate.Execute(&sqlByteBuffer, nil)
	if err != nil {
		return nil
	}

	return &internalmodels.TestConfig{
		TestName:      refName,
		GoName:        goFuncName,
		NameUpperCase: strings.ToUpper(fmt.Sprintf("%s_%s", stage, strings.ReplaceAll(fileName, ".sql", ""))),
		SqlByteBuffer: sqlByteBuffer,
		TestProfile:   globalTestProfile,
	}
}
