package ironman

import (
	"bytes"
	"context"

	"log"
	"os"
	"path/filepath"
	gtemplate "text/template"

	"github.com/ironman-project/ironman/template"
	"github.com/ironman-project/ironman/template/engine"
	"github.com/ironman-project/ironman/template/engine/goengine"
	"github.com/ironman-project/ironman/template/values"

	"github.com/ironman-project/ironman/template/validator"

	"github.com/blevesearch/bleve"
	"github.com/ironman-project/ironman/template/manager"
	"github.com/ironman-project/ironman/template/manager/git"
	"github.com/ironman-project/ironman/template/model"
	"github.com/ironman-project/ironman/template/repository"
	brepository "github.com/ironman-project/ironman/template/repository/bleve"
	"github.com/pkg/errors"
)

const (
	indexName          = "templates.index"
	templatesDirectory = "templates"
)

var validationTempl *gtemplate.Template

const validatoinTemplateText = ``

func init() {
	var err error
	validationTempl, err = gtemplate.New("validationTemplate").Parse(validatoinTemplateText)
	if err != nil {
		log.Fatalf("Failed to initialize validation errors template %s", err)
	}
}

//Ironman is the one administering the local
type Ironman struct {
	manager     manager.Manager
	modelReader model.Reader
	repository  repository.Repository
	home        string
	validators  []validator.Validator
}

//New returns a new instance of ironman
func New(home string, options ...Option) *Ironman {

	ir := &Ironman{home: home}
	for _, option := range options {
		option(ir)
	}
	if ir.manager == nil {
		manager := git.New(home, templatesDirectory)
		ir.manager = manager
	}

	if ir.repository == nil {
		indexPath := filepath.Join(home, indexName)
		index, err := buildIndex(indexPath)
		if err != nil {
			log.Fatal("Failed to create ironman templates index", err)
		}
		ir.repository = brepository.New(
			brepository.SetIndex(index),
		)
	}

	if ir.modelReader == nil {
		decoder := model.NewDecoder(model.DecoderTypeYAML)
		modelReader := model.NewFSReader([]string{".git"}, model.MetadataFileExtensionYAML, decoder)
		ir.modelReader = modelReader
	}

	if ir.validators == nil {
		ir.validators = []validator.Validator{}
	}

	return ir
}

func buildIndex(path string) (bleve.Index, error) {
	// open the index
	index, err := bleve.Open(path)
	if err == bleve.ErrorIndexPathDoesNotExist {
		index, err = brepository.BuildIndex(path)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	return index, nil
}

//Install installs a new template based on a template locator
func (i *Ironman) Install(templateLocator string) error {

	ID, err := i.manager.Install(templateLocator)

	if err != nil {
		return err
	}

	templatePath := i.manager.TemplateLocation(ID)

	model, err := i.modelReader.Read(templatePath)

	if err != nil {
		return err
	}

	//validate model
	for _, validator := range i.validators {
		valid, validationErr, err := validator.Validate(model)

		if err != nil {
			return errors.Wrap(err, "Failed to validate model")
		}

		if !valid {
			var validationErrBuffer bytes.Buffer
			err := validationTempl.Execute(&validationErrBuffer, validationErr)

			if err != nil {
				return errors.Wrap(err, "Failed to create validation error message")
			}

			return errors.New(validationErrBuffer.String())
		}
	}

	_, err = i.repository.Index(model)

	if err != nil {
		//rollback manager installation
		_ = i.manager.Uninstall(ID)
		return err
	}

	return nil
}

//Link Creates a symlink to the ironman repository from any path in the filesystem
func (i *Ironman) Link(templatePath, templateID string) error {

	err := i.manager.Link(templatePath, templateID)

	if err != nil {
		return err
	}

	model, err := i.modelReader.Read(templatePath)

	if err != nil {
		return err
	}

	_, err = i.repository.Index(model)

	if err != nil {
		return err
	}

	return nil
}

//List returns a list of all the installed ironman templates
func (i *Ironman) List() ([]*model.Template, error) {
	results, err := i.repository.List()
	if err != nil {
		return nil, err
	}

	return results, nil
}

//Uninstall uninstalls an ironman template
func (i *Ironman) Uninstall(templateID string) error {

	exists, err := i.repository.Exists(templateID)

	if err != nil {
		return errors.Wrapf(err, "Failed to validate if template exists %s", templateID)
	}

	if !exists {
		return errors.Errorf("Template is not installed")
	}

	err = i.manager.Uninstall(templateID)

	if err != nil {
		return err
	}

	_, err = i.repository.Delete(templateID)

	if err != nil {
		return err
	}

	return nil
}

//Unlink unlinks a previously linked ironman template
func (i *Ironman) Unlink(templateID string) error {

	err := i.manager.Unlink(templateID)

	if err != nil {
		return err
	}

	_, err = i.repository.Delete(templateID)

	if err != nil {
		return err
	}

	return nil
}

//Update updates an iroman template
func (i *Ironman) Update(templateID string) error {
	exists, err := i.repository.Exists(templateID)

	if err != nil {
		return errors.Wrapf(err, "Failed to validate if template exists %s", templateID)
	}

	if !exists {
		return errors.Errorf("Template is not installed")
	}

	err = i.manager.Update(templateID)

	if err != nil {
		return err
	}

	return nil
}

//Generate generates a new file or directory based on a generator
func (i *Ironman) Generate(context context.Context, templateID string, generatorID string, generationPath string, values values.Values) error {

	exists, err := i.repository.Exists(templateID)

	if err != nil {
		return errors.Wrapf(err, "Failed to validate if template exists %s", templateID)
	}

	if !exists {
		return errors.Errorf("Template is not installed")
	}

	templateModel, err := i.repository.FindTemplateByID(templateID)

	if err != nil {
		return errors.Wrapf(err, "Could not find template by ID %s", templateID)
	}

	genteratorModel := templateModel.Generator(generatorID)

	if genteratorModel == nil {
		return errors.Errorf("Generator %s does not exists", generatorID)
	}

	generatorPath := filepath.Join(i.home, templatesDirectory, templateModel.DirectoryName, genteratorModel.DirectoryName)

	data := template.GeneratorData{
		Template:  templateModel,
		Generator: genteratorModel,
		Values:    values,
	}

	engineFactory := func() engine.Engine {
		return goengine.New("ironman")
	}

	generator := template.NewGenerator(generatorPath, generationPath, []string{".ironman.yaml"}, data, engineFactory)

	if err := generator.Generate(context); err != nil {
		return err
	}

	return nil
}

//InitIronmanHome inits the ironman home directory
func InitIronmanHome(ironmanHome string) error {
	if _, err := os.Stat(ironmanHome); os.IsNotExist(err) {
		err := os.Mkdir(ironmanHome, os.ModePerm)
		if err != nil {
			return errors.Wrap(err, "Failed to initialize ironman home")
		}

		err = os.Mkdir(filepath.Join(ironmanHome, templatesDirectory), os.ModePerm)

		if err != nil {
			return errors.Wrap(err, "Failed to initialize ironman home")
		}
	}
	return nil
}
