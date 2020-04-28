package ezactions

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

// Main is to be used in main(). Main will run the action.
// If the -generate flag is set, it will generate action.yml and Dockerfile instead.
func (a *Action) Main() {
	var generate bool
	flag.BoolVar(&generate, "generate", false, "generate action.yml")
	flag.Parse()
	if generate {
		err := a.generate()
		if err != nil {
			log.Fatalf("error generating: %v", err)
		}
		return
	}
	runAction(*a)
}

// Action is the definition for an action.
type Action struct {
	// **Required** The name of your action. GitHub displays the name in the Actions tab to help visually identify
	// actions in each job.
	//
	// https://help.github.com/en/actions/building-actions/metadata-syntax-for-github-actions#name
	Name string

	// **Optional** The name of the action's author.
	//
	// https://help.github.com/en/actions/building-actions/metadata-syntax-for-github-actions#author
	Author string

	// **Required** A short description of the action.
	//
	// https://help.github.com/en/actions/building-actions/metadata-syntax-for-github-actions#description
	Description string

	// **Optional** Input parameters allow you to specify data that the action expects to use during runtime.
	//  Input ids with uppercase letters are converted to lowercase during runtime. GitHub recommendeds using lowercase
	// input ids.
	//
	// https://help.github.com/en/actions/building-actions/metadata-syntax-for-github-actions#inputs
	Inputs []ActionInput

	// **Optional** Output parameters allow you to declare data that an action sets. Actions that run later in a
	// workflow can use the output data set in previously run actions. For example, if you had an action that performed
	// the addition of two inputs (x + y = z), the action could output the sum (z) for other actions to use as an input.
	//
	// https://help.github.com/en/actions/building-actions/metadata-syntax-for-github-actions#outputs
	Outputs []ActionOutput

	// **Required** The code that your action runs. Declared inputs are available in the inputs parameter.
	// Returned outputs are available to subsequent workflow steps.
	Run func(inputs map[string]string, resources *RunResources) (outputs map[string]string, err error)
}

// ActionInput is an input declared for an action.
type ActionInput struct {

	// **Required** A string identifier to associate with the input. The value of <input_id> is a map of the input's
	// metadata. The <input_id> must be a unique identifier within the inputs object. The <input_id> must start with a
	// letter or _ and contain only alphanumeric characters, -, or _.
	//
	// https://help.github.com/en/actions/building-actions/metadata-syntax-for-github-actions#inputsinput_id
	ID string

	// **Required** A string description of the input parameter.
	//
	// https://help.github.com/en/actions/building-actions/metadata-syntax-for-github-actions#inputsinput_iddescription
	Description string

	// **Required** A boolean to indicate whether the action requires the input parameter. Set to true when the
	// parameter is required.
	//
	// https://help.github.com/en/actions/building-actions/metadata-syntax-for-github-actions#inputsinput_idrequired
	Required bool

	// **Optional** A string representing the default value. The default value is used when an input parameter isn't
	// specified in a workflow file.
	//
	// https://help.github.com/en/actions/building-actions/metadata-syntax-for-github-actions#inputsinput_iddefault
	Default string
}

// ActionOutput is an output to be declared in an action.yml
type ActionOutput struct {

	// **Required** A string identifier to associate with the output. The value of <output_id> is a map of the output's
	// metadata. The <output_id> must be a unique identifier within the outputs object. The <output_id> must start with
	// a letter or _ and contain only alphanumeric characters, -, or _.
	//
	// https://help.github.com/en/actions/building-actions/metadata-syntax-for-github-actions#outputsoutput_id
	ID string

	// **Required** A string description of the output parameter.
	//
	// https://help.github.com/en/actions/building-actions/metadata-syntax-for-github-actions#outputsoutput_iddescription
	Description string
}

// RunResources contains resources available to action.Run()
type RunResources struct {
	Action            *Action            // The action being called. Just in case you need to inspect it.
	WorkflowCommander *WorkflowCommander // So you can issue workflow commands.
}

func (a *Action) generate() error {
	actionfile, err := os.Create("action.yml")
	if err != nil {
		return err
	}
	err = writeActionYML(actionfile, a)
	if err != nil {
		return err
	}

	dockerfile, err := os.Create("Dockerfile")
	if err != nil {
		return err
	}
	err = writeDockerFile(dockerfile, nil)
	if err != nil {
		return err
	}
	return nil
}

func runAction(action Action) {
	commander := &WorkflowCommander{
		Printer: func(s string) {
			fmt.Print(s)
		},
	}

	inputs := make(map[string]string, len(action.Inputs))
	for _, input := range action.Inputs {
		envVar := "INPUT_" + strings.ToUpper(strings.ReplaceAll(input.ID, " ", "_"))
		val, ok := os.LookupEnv(envVar)
		if ok {
			inputs[input.ID] = val
		}
	}
	got, err := action.Run(inputs, &RunResources{
		Action:            &action,
		WorkflowCommander: commander,
	})

	// Intentionally not checking for an error because we want to set any outputs
	// returned even in the case of an error
	for key, val := range got {
		commander.SetOutputParameter(key, val)
	}

	if err != nil {
		commander.SetErrorMessage(err.Error(), nil)
		os.Exit(1)
	}
}
