# Initial Design Doc 

There has been a significant amount of changes to the code base so this is not meant to reflect the contents of the codebase. Instead this should serve as an introduction into how Golang Interfaces and Structs were used to make the initial iteration of Acts/Blocks. An important note to make is that this structure is meant to emulate a compiler in the tokenization and parsing of these TTPs.

## Act 

This structure contains a high level block, designated an Act to indicate its part in the overall execution of the runbook/TTP.

Every sub structure may embed the Act which will dictate common details shared across all Acts in the codebase. Additionally, sub structures may redefine these fields outside of the embedding to offer up an override. 

## Step

Steps contains a high level interface which, if desired by the developer, may be implemented in any sub structs to indicate to the compiler that these sub structs are also to be considered a Step. Using this system we can call methods on a Step without knowing its specific type. This enables us to create highly typed objects that share high level functionality in order to operate together.

```go
type Act struct {
	Condition   string            `yaml:"if,omitempty"`
	Environment map[string]string `yaml:"env,flow,omitempty"`
	Name        string            `yaml:"name"`
	WorkDir     string            `yaml:"-"`
	Type        StepType          `yaml:"-"`
	success     bool
	stepRef     map[string]Step
	output      map[string]any
}

type Step interface {
	Setup(env map[string]string, outputRef map[string]Step)
	SetDir(dir string)
	// Need list in case some steps are encapsulating many cleanup steps
	GetCleanup() []CleanupAct
	// Execute will need to take care of the condition checks/etc...
	Execute() error
	IsNil() bool
	ExplainInvalid() error
	Validate() error
	FetchArgs(args []string) []string
	GetOutput() map[string]any
	SearchOutput(arg string) string
	SetOutputSuccess(output *bytes.Buffer, exit int)
	Success() bool
	StepName() string
	GetType() StepType
}

```

## Expanding on Step/Act

To highlight the usefulness of tokenization with this setup we can create the sub structures which will make up our basic language. 

1. BasicBlock (runs interpreter bash/sh/powershell)
2. FileBlock (runs a program with default method provided)

We can build our language model like such:

a TTP is made up of a collection of BasicBlocks and FileBlocks.

```
RunBook = (TTP)*
TTP = (Basic|File)*
```

That is pretty simple right?

So what this basic setup allows us to do is define our language which is simple for now, but can become more complex. Now we just need a way to delineate between the types. 

This is where the interface comes in.

We can leverage Unmarshalling to determine the tokens defined at each step by validating that they unpack into the correct structure like so:

```go

pseudo code

func unmarshal {
	partial = Partial Structure
	partial = unmarshal(node)


	for step in partial.steps {
		a = unmarshal(step)
		if a == Basic {
			...
		}
		else if a == File {
			...
		}
	}
}

```


## Complexity

You can now extend this by defining more bits of the language like so:

1. BasicBlock (runs interpreter bash/sh/powershell)
2. FileBlock (runs a program with default method provided)
3. SubTTPs (file which contains collections of Basic and File)

The language is now:

```
RunBook = (TTP|SubTTP)*
TTP = (Basic|File)*
SubTTP = (TTP)*
```

SubTTP contain a list of TTPs which expand into our subset of Basic and File.

This means when we unmarshal a SubTTP we can bottom out, otherwise we would infinitely recurse as we unpack.

```
type SubTTPStep struct {
	*Act       `yaml:",inline"`
	TtpFile    string    `yaml:"ttp"`
	FileSystem fs.StatFS `yaml:"-,omitempty"`
	CleanupSteps []CleanupAct `yaml:"-,omitempty"`
	ttp          TTP
}

```

So you can see the TTP embedded in the struct is what prevents the SubTTPStep from causing recusrion issues.

## Cleanup

Cleanup is an embedded structure for all Acts to make use of. 

For Basic and File blocks they each get their own Cleanup step. While the SubTTP contains a list of CleanupSteps which are gathered as execution of each TTP in the TTP structure.



## Resources

- [LogRocket Structs/Interfaces](https://blog.logrocket.com/exploring-structs-interfaces-go/)
- [GobyExample Interfaces](https://gobyexample.com/interfaces)
- [Go Dev Tour](https://go.dev/tour/methods/9)
- [Tokenizers and Compilers](https://www.cs.man.ac.uk/~pjj/farrell/comp3.html)
