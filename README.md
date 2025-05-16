# Terraform Loading Bar

A Go tool that provides a visual progress bar for Terraform operations by parsing the JSON output stream. It offers real-time feedback during both planning and apply phases of Terraform operations.

## Features

- Real-time progress updates with step-by-step tracking
- Visual progress bar showing completion status
- Support for both planning and apply phases
- Resource operation status tracking (create, modify, delete)
- Error highlighting and display
- Clean, formatted output
- JSON stream parsing for accurate progress tracking
- Seamless integration with Let'em Cook! (LEMC) for browser-based progress display

## Using with Let'em Cook! (LEMC)

The progress bar can be integrated with [Let'em Cook! (LEMC)](https://github.com/jaredfolkins/letemcook) to stream the UI to the browser. Here's how to use the ProgressHandler with LEMC:

```go
package main

import (
    "fmt"
    "io"
    "os"
    "os/exec"
    "strings"
    "github.com/jaredfolkins/terraform-loading-bar/progress"
)

func main() {
    // Initialize Terraform
    initCmd := exec.Command("terraform", "init")
    initCmd.Stdout = os.Stdout
    initCmd.Stderr = os.Stderr
    if err := initCmd.Run(); err != nil {
        panic(err)
    }

    // Create plan
    planCmd := exec.Command("terraform", "plan", "-out", "terraform.plan")
    planCmd.Stdout = os.Stdout
    planCmd.Stderr = os.Stderr
    if err := planCmd.Run(); err != nil {
        panic(err)
    }

    // Apply with JSON output
    applyCmd := exec.Command("terraform", "apply", "-json", "-auto-approve", "terraform.plan")
    applyOutput, err := applyCmd.StdoutPipe()
    if err != nil {
        panic(err)
    }
    applyCmd.Stderr = os.Stderr

    // Start the apply command
    if err := applyCmd.Start(); err != nil {
        panic(err)
    }

    // Create a progress handler
    progressHandler := progress.NewProgressHandler(applyOutput)

    // Stream the output in real-time with progress bar
    for {
        line, err := progressHandler.ReadLine()
        if err != nil {
            if err == io.EOF {
                break
            }
            panic(err)
        }

        if line != "" {
            // Use LEMC's trunc verb to update the progress bar in place
            fmt.Printf("lemc.html.trunc; %s\n", strings.TrimSpace(line))
        }
    }

    // Wait for the apply command to complete
    if err := applyCmd.Wait(); err != nil {
        panic(err)
    }
}
```

### How It Works with LEMC

1. The `ProgressHandler` processes Terraform's JSON output stream in real-time
2. Each line of progress is formatted with a visual progress bar
3. Using LEMC's `lemc.html.trunc` verb, each line replaces the previous one in the browser
4. This creates a smooth, animated progress bar showing:
   - Current operation number and total operations
   - Visual progress bar
   - Resource name and current operation
   - Error messages when they occur

### Progress Bar Format

The progress bar displays in the following format:
```
(current_step)[====================](total_steps) resource_name: operation...
```

For example:
```
(01)[====================](18) google_compute_network.vpc: Creating...
```

During the planning phase, it shows:
```
[     PLANNING     ] Planning...
```

## Installation

```bash
go install github.com/jaredfolkins/terraform-loading-bar@latest
```

## Usage

### Basic Workflow

1. Initialize your Terraform configuration:
```bash
terraform init
```

2. Create a plan file:
```bash
terraform plan -out terraform.plan
```

3. Apply the plan with JSON output and pipe it through the loading bar:
```bash
terraform apply -json terraform.plan | terraform-loading-bar
```

### Using in Go Code

You can also use the parser directly in your Go application to stream Terraform output:

```go
package main

import (
    "os"
    "os/exec"
    "github.com/jaredfolkins/terraform-loading-bar/progress"
)

func main() {
    // Initialize Terraform
    initCmd := exec.Command("terraform", "init")
    initCmd.Stdout = os.Stdout
    initCmd.Stderr = os.Stderr
    if err := initCmd.Run(); err != nil {
        panic(err)
    }

    // Create plan
    planCmd := exec.Command("terraform", "plan", "-out", "terraform.plan")
    planCmd.Stdout = os.Stdout
    planCmd.Stderr = os.Stderr
    if err := planCmd.Run(); err != nil {
        panic(err)
    }

    // Apply with JSON output
    applyCmd := exec.Command("terraform", "apply", "-json", "-auto-approve", "terraform.plan")
    applyOutput, err := applyCmd.StdoutPipe()
    if err != nil {
        panic(err)
    }
    applyCmd.Stderr = os.Stderr

    // Start the apply command
    if err := applyCmd.Start(); err != nil {
        panic(err)
    }

    // Stream the output through the progress bar
    if err := progress.ProcessJSONStream(applyOutput, os.Stdout); err != nil {
        panic(err)
    }

    // Wait for the apply command to complete
    if err := applyCmd.Wait(); err != nil {
        panic(err)
    }
}
```

### Customizing Message Trimming

By default, long messages are truncated to keep the progress bar output on a
single line. You can adjust how many characters are retained before the
ellipsis is added:

```go
progress.SetTrimLength(60) // allow up to 60 characters
```

### Getting Progress Output as String

If you want to get the progress output as a string instead of printing it directly, you can use the `GetProgressOutput` function:

```go
package main

import (
    "os"
    "os/exec"
    "fmt"
    "github.com/jaredfolkins/terraform-loading-bar/progress"
)

func main() {
    // Initialize Terraform
    initCmd := exec.Command("terraform", "init")
    initCmd.Stdout = os.Stdout
    initCmd.Stderr = os.Stderr
    if err := initCmd.Run(); err != nil {
        panic(err)
    }

    // Create plan
    planCmd := exec.Command("terraform", "plan", "-out", "terraform.plan")
    planCmd.Stdout = os.Stdout
    planCmd.Stderr = os.Stderr
    if err := planCmd.Run(); err != nil {
        panic(err)
    }

    // Apply with JSON output
    applyCmd := exec.Command("terraform", "apply", "-json", "-auto-approve", "terraform.plan")
    applyOutput, err := applyCmd.StdoutPipe()
    if err != nil {
        panic(err)
    }
    applyCmd.Stderr = os.Stderr

    // Start the apply command
    if err := applyCmd.Start(); err != nil {
        panic(err)
    }

    // Get the progress output as a string
    output, err := progress.GetProgressOutput(applyOutput)
    if err != nil {
        panic(err)
    }

    // Print the output yourself
    fmt.Print(output)

    // Wait for the apply command to complete
    if err := applyCmd.Wait(); err != nil {
        panic(err)
    }
}
```

## Error Handling

The tool will display any Terraform errors in a clear format:
```
TERRAFORM ERROR: error message
```

## Development

### Building from Source

```bash
git clone https://github.com/jaredfolkins/terraform-loading-bar.git
cd terraform-loading-bar
go build
```

### Running Tests

```bash
go test ./...
```

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. 