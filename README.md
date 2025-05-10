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

## Installation

```bash
go install github.com/jaredfolkins/terraform-loading-bar@latest
```

## Usage

The tool reads Terraform's JSON output stream and displays a progress bar showing the status of resource creation, modification, or deletion.

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
    applyCmd := exec.Command("terraform", "apply", "-json", "terraform.plan")
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
    applyCmd := exec.Command("terraform", "apply", "-json", "terraform.plan")
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

This example shows how to:
1. Initialize Terraform
2. Create a plan
3. Apply the plan with JSON output
4. Get the progress output as a string and print it yourself

The progress bar will show real-time updates as resources are created, modified, or deleted.

### Output Format

The tool displays progress in the following format:
```
(current_step)[====================](total_steps) resource_name: operation...
```

For example:
```
(1)[====================](18) google_compute_network.vpc: Creating...
```

Where:
- `current_step`: Current operation number
- `====================`: Visual progress bar
- `total_steps`: Total number of operations
- `resource_name`: Name of the resource being operated on
- `operation`: Current operation (Creating, Modifying, Deleting, etc.)

During the planning phase, the tool displays:
```
[     PLANNING     ] Planning...
```

### Error Handling

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