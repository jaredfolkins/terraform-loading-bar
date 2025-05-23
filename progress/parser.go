package progress

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

// TrimLength controls how many characters of the message are retained
// before being truncated with an ellipsis. The default is 48.
var TrimLength = 48

// SetTrimLength allows callers to change the message trimming length.
// Values under 4 are ignored to ensure space for the ellipsis.
func SetTrimLength(n int) {
	if n >= 4 {
		TrimLength = n
	}
}

// TerraformLogEntry represents a generic log entry from Terraform JSON output.
type TerraformLogEntry struct {
	Level      string                 `json:"@level"`
	Message    string                 `json:"@message"`
	Module     string                 `json:"@module"`
	Timestamp  string                 `json:"@timestamp"`
	Type       string                 `json:"type"`
	Hook       *HookData              `json:"hook,omitempty"`
	Change     *ChangeData            `json:"change,omitempty"`
	Changes    *ChangesSummary        `json:"changes,omitempty"`
	Outputs    map[string]OutputEntry `json:"outputs,omitempty"`    // To capture the final outputs block
	Terraform  string                 `json:"terraform,omitempty"`  // For version info
	UI         string                 `json:"ui,omitempty"`         // For version info
	Diagnostic *DiagnosticData        `json:"diagnostic,omitempty"` // For warnings and errors
}

// HookData contains information about the resource being acted upon.
type HookData struct {
	Resource       ResourceInfo `json:"resource"`
	Action         string       `json:"action,omitempty"`
	IDKey          string       `json:"id_key,omitempty"`
	IDValue        string       `json:"id_value,omitempty"`
	ElapsedSeconds float64      `json:"elapsed_seconds,omitempty"`
}

// ChangeData also contains information about planned changes to a resource.
type ChangeData struct {
	Resource ResourceInfo `json:"resource"`
	Action   string       `json:"action"`
}

// ResourceInfo identifies a specific Terraform resource.
type ResourceInfo struct {
	Addr            string `json:"addr"`
	Module          string `json:"module"`
	Resource        string `json:"resource"`
	ImpliedProvider string `json:"implied_provider"`
	ResourceType    string `json:"resource_type"`
	ResourceName    string `json:"resource_name"`
	ResourceKey     string `json:"resource_key"`
}

// ChangesSummary provides a summary of planned changes.
type ChangesSummary struct {
	Add       int    `json:"add"`
	Change    int    `json:"change"`
	Remove    int    `json:"remove"`
	Import    int    `json:"import"` // Though not in example, good to have
	Operation string `json:"operation"`
}

// OutputEntry represents a single output value.
// We might not use this directly for the progress bar, but it's good for completeness.
type OutputEntry struct {
	Sensitive bool        `json:"sensitive"`
	Type      interface{} `json:"type"` // Type can be a string or a more complex structure
	Value     interface{} `json:"value,omitempty"`
	Action    string      `json:"action,omitempty"` // For planned outputs
}

// DiagnosticData contains information about warnings and errors.
type DiagnosticData struct {
	Severity string `json:"severity"`
	Summary  string `json:"summary"`
	Detail   string `json:"detail"`
}

// ProgressHandler handles streaming progress for Terraform operations
type ProgressHandler struct {
	reader           io.Reader
	scanner          *bufio.Scanner
	totalSteps       int
	currentStep      int
	progressBarWidth int
	isPlanning       bool
	lastFullMessage  string
	resourceMessages map[string]string
	lineChan         chan string
	errChan          chan error
	originalOutput   []string // Buffer to store original JSON lines
}

// GetOriginalOutput returns the buffered original JSON output
func (ph *ProgressHandler) GetOriginalOutput() []string {
	return ph.originalOutput
}

// NewProgressHandler creates a new ProgressHandler for the given reader
func NewProgressHandler(reader io.Reader) *ProgressHandler {
	ph := &ProgressHandler{
		reader:           reader,
		scanner:          bufio.NewScanner(reader),
		totalSteps:       0,
		currentStep:      0,
		progressBarWidth: 0,
		isPlanning:       true,
		lastFullMessage:  "Planning...",
		resourceMessages: make(map[string]string),
		lineChan:         make(chan string),
		errChan:          make(chan error),
		originalOutput:   make([]string, 0),
	}

	// Start processing in a goroutine
	go ph.process()

	return ph
}

// ReadLine reads the next line of progress output
func (ph *ProgressHandler) ReadLine() (string, error) {
	select {
	case line := <-ph.lineChan:
		return line, nil
	case err := <-ph.errChan:
		return "", err
	}
}

// process handles the actual processing of the JSON stream
func (ph *ProgressHandler) process() {
	defer close(ph.lineChan)
	defer close(ph.errChan)

	// Send initial progress bar
	ph.lineChan <- getProgressString(ph.currentStep, ph.totalSteps, ph.progressBarWidth, ph.lastFullMessage, ph.isPlanning)

	// Check for scanner errors before the main loop
	if err := ph.scanner.Err(); err != nil {
		ph.errChan <- err
		return
	}

	for ph.scanner.Scan() {
		line := ph.scanner.Text()
		// Store the original line in the buffer
		ph.originalOutput = append(ph.originalOutput, line)

		var entry TerraformLogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			// Non-JSON lines might be progress updates from plugins, or other info.
			// We can choose to send them through or log them for debugging.
			// For now, let's assume they are not part of the progress bar data itself.
			continue
		}

		msg := entry.Message
		currentResourceAddr := ""

		if entry.Hook != nil && entry.Hook.Resource.Addr != "" {
			currentResourceAddr = entry.Hook.Resource.Addr
		}

		// Update totalSteps if a plan summary appears
		if entry.Type == "change_summary" && entry.Changes != nil && entry.Changes.Operation == "plan" {
			// Double the total steps to account for both start and completion events
			newTotal := (entry.Changes.Add + entry.Changes.Change + entry.Changes.Remove) * 2
			if ph.totalSteps == 0 || newTotal > ph.totalSteps {
				ph.totalSteps = newTotal
				ph.progressBarWidth = ph.totalSteps
			}
			ph.isPlanning = false
			ph.lastFullMessage = msg
			ph.lineChan <- getProgressString(ph.currentStep, ph.totalSteps, ph.progressBarWidth, ph.lastFullMessage, ph.isPlanning)
			continue
		}

		// Set isPlanning to false when we see the first apply_start
		if entry.Type == "apply_start" {
			ph.isPlanning = false // Definitely not planning anymore

			// If totalSteps is still 0, try to derive it from planned_change events in the buffer
			if ph.totalSteps == 0 && len(ph.originalOutput) > 0 {
				plannedResources := make(map[string]bool)
				for _, bufferedLine := range ph.originalOutput {
					var bufferedEntry TerraformLogEntry
					if json.Unmarshal([]byte(bufferedLine), &bufferedEntry) == nil {
						if bufferedEntry.Type == "planned_change" && bufferedEntry.Change != nil && bufferedEntry.Change.Resource.Addr != "" {
							plannedResources[bufferedEntry.Change.Resource.Addr] = true
						}
					}
				}
				if len(plannedResources) > 0 {
					ph.totalSteps = len(plannedResources) * 2
					ph.progressBarWidth = ph.totalSteps
				}
			}

			if ph.currentStep < ph.totalSteps {
				ph.currentStep++
			}
			if currentResourceAddr != "" {
				ph.resourceMessages[currentResourceAddr] = msg
				ph.lastFullMessage = msg
			} else {
				ph.lastFullMessage = msg
			}
		} else if entry.Type == "apply_complete" {
			if ph.currentStep < ph.totalSteps {
				ph.currentStep++
			}
			if currentResourceAddr != "" {
				ph.resourceMessages[currentResourceAddr] = msg
				ph.lastFullMessage = msg
			} else {
				ph.lastFullMessage = msg
			}
		} else if entry.Type == "apply_progress" {
			// Don't increment step, just update message
			if currentResourceAddr != "" {
				ph.resourceMessages[currentResourceAddr] = msg
				ph.lastFullMessage = msg
			} else {
				ph.lastFullMessage = msg
			}
		} else if entry.Type == "change_summary" && entry.Changes != nil && entry.Changes.Operation == "apply" {
			// Instead of jumping to total, increment until we reach it
			for ph.currentStep < ph.totalSteps {
				ph.currentStep++
				// Show "Finalizing..." message during completion
				ph.lastFullMessage = "Finalizing..."
				ph.lineChan <- getProgressString(ph.currentStep, ph.totalSteps, ph.progressBarWidth, ph.lastFullMessage, ph.isPlanning)
				// Add a small delay to make the progress visible
				time.Sleep(50 * time.Millisecond)
			}
			ph.lastFullMessage = msg
		} else if entry.Type == "outputs" {
			ph.lastFullMessage = "Processing outputs..."
		} else if msg != "" && entry.Level == "error" {
			ph.lineChan <- fmt.Sprintf("TERRAFORM ERROR: %s", msg)
			ph.lastFullMessage = msg
		} else if msg != "" {
			ph.lastFullMessage = msg
		}

		ph.lineChan <- getProgressString(ph.currentStep, ph.totalSteps, ph.progressBarWidth, ph.lastFullMessage, ph.isPlanning)

		if entry.Type == "change_summary" && entry.Changes != nil && entry.Changes.Operation == "apply" {
			ph.lineChan <- ""
		}
	}

	if err := ph.scanner.Err(); err != nil {
		ph.errChan <- err
		return
	}

	// Signal EOF after processing all output
	ph.errChan <- io.EOF
}

// GetProgressOutput reads Terraform JSON output from reader and returns the progress bar output as a string.
// This function is similar to ProcessJSONStream but returns the output instead of printing it.
func GetProgressOutput(reader io.Reader) (string, error) {
	scanner := bufio.NewScanner(reader)
	totalSteps := 0
	currentStep := 0
	progressBarWidth := 0            // Will be calculated based on total steps
	isPlanning := true               // Track if we're in planning phase
	lastFullMessage := "Planning..." // Set initial message
	var output strings.Builder

	// First pass to estimate total steps from the plan summary or count resources
	var lines []string
	seeker, isSeeker := reader.(io.Seeker)
	if isSeeker {
		initialPos, _ := seeker.Seek(0, io.SeekCurrent)
		scannerForCount := bufio.NewScanner(reader)
		for scannerForCount.Scan() {
			line := scannerForCount.Text()
			lines = append(lines, line)
			var entry TerraformLogEntry
			if err := json.Unmarshal([]byte(line), &entry); err == nil {
				if entry.Type == "change_summary" && entry.Changes != nil && entry.Changes.Operation == "plan" {
					// Double the total steps to account for both start and completion events
					totalSteps = (entry.Changes.Add + entry.Changes.Change + entry.Changes.Remove) * 2
					progressBarWidth = totalSteps // Set bar width equal to total steps
					isPlanning = false            // We're done planning
					break                         // Found the plan summary
				}
			}
		}
		seeker.Seek(initialPos, io.SeekStart) // Reset reader for the main processing pass
		scanner = bufio.NewScanner(reader)    // Re-initialize scanner
	}

	// If totalSteps is still 0 after the first pass (or if it wasn't a seeker),
	// we can try to count distinct resources from planned_change as a fallback.
	if totalSteps == 0 && len(lines) > 0 {
		plannedResources := make(map[string]bool)
		for _, line := range lines {
			var entry TerraformLogEntry
			if err := json.Unmarshal([]byte(line), &entry); err == nil {
				if entry.Type == "planned_change" && entry.Change != nil && entry.Change.Resource.Addr != "" {
					plannedResources[entry.Change.Resource.Addr] = true
				} else if entry.Type == "change_summary" && entry.Changes != nil && entry.Changes.Operation == "plan" {
					// Double the total steps to account for both start and completion events
					totalSteps = (entry.Changes.Add + entry.Changes.Change + entry.Changes.Remove) * 2
					progressBarWidth = totalSteps // Set bar width equal to total steps
					isPlanning = false            // We're done planning
					break
				}
			}
		}
		if totalSteps == 0 {
			// Double the total steps to account for both start and completion events
			totalSteps = len(plannedResources) * 2
			progressBarWidth = totalSteps // Set bar width equal to total steps
		}
		scanner = bufio.NewScanner(strings.NewReader(strings.Join(lines, "\n")))
	}

	// Get initial progress bar
	output.WriteString(getProgressString(currentStep, totalSteps, progressBarWidth, lastFullMessage, isPlanning))
	output.WriteString("\n")

	resourceMessages := make(map[string]string)

	for scanner.Scan() {
		line := scanner.Text()
		var entry TerraformLogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}

		msg := entry.Message
		currentResourceAddr := ""

		if entry.Hook != nil && entry.Hook.Resource.Addr != "" {
			currentResourceAddr = entry.Hook.Resource.Addr
		}

		// Update totalSteps if a plan summary appears mid-stream
		if entry.Type == "change_summary" && entry.Changes != nil && entry.Changes.Operation == "plan" {
			// Double the total steps to account for both start and completion events
			newTotal := (entry.Changes.Add + entry.Changes.Change + entry.Changes.Remove) * 2
			if totalSteps == 0 || newTotal > totalSteps {
				totalSteps = newTotal
				progressBarWidth = totalSteps // Update bar width when total steps changes
			}
			isPlanning = false // We're done planning
			lastFullMessage = msg
			output.WriteString(getProgressString(currentStep, totalSteps, progressBarWidth, lastFullMessage, isPlanning))
			output.WriteString("\n")
			continue
		}

		if entry.Type == "apply_start" {
			if currentStep < totalSteps {
				currentStep++
			}
			if currentResourceAddr != "" {
				resourceMessages[currentResourceAddr] = msg
				lastFullMessage = msg
			} else {
				lastFullMessage = msg
			}
		} else if entry.Type == "apply_complete" {
			if currentStep < totalSteps {
				currentStep++
			}
			if currentResourceAddr != "" {
				resourceMessages[currentResourceAddr] = msg
				lastFullMessage = msg
			} else {
				lastFullMessage = msg
			}
		} else if entry.Type == "apply_progress" {
			// Don't increment step, just update message
			if currentResourceAddr != "" {
				resourceMessages[currentResourceAddr] = msg
				lastFullMessage = msg
			} else {
				lastFullMessage = msg
			}
		} else if entry.Type == "change_summary" && entry.Changes != nil && entry.Changes.Operation == "apply" {
			// Instead of jumping to total, increment until we reach it
			for currentStep < totalSteps {
				currentStep++
				// Show "Finalizing..." message during completion
				lastFullMessage = "Finalizing..."
				output.WriteString(getProgressString(currentStep, totalSteps, progressBarWidth, lastFullMessage, isPlanning))
				output.WriteString("\n")
				// Add a small delay to make the progress visible
				time.Sleep(50 * time.Millisecond)
			}
			lastFullMessage = msg
		} else if entry.Type == "outputs" {
			lastFullMessage = "Processing outputs..."
		} else if msg != "" && entry.Level == "error" {
			output.WriteString(fmt.Sprintf("TERRAFORM ERROR: %s\n", msg))
			lastFullMessage = msg
		} else if msg != "" {
			lastFullMessage = msg
		}

		output.WriteString(getProgressString(currentStep, totalSteps, progressBarWidth, lastFullMessage, isPlanning))
		output.WriteString("\n")

		if entry.Type == "change_summary" && entry.Changes != nil && entry.Changes.Operation == "apply" {
			output.WriteString("\n")
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading input: %w", err)
	}

	return output.String(), nil
}

// getProgressString returns the progress bar string without printing it
func getProgressString(current, total, width int, message string, isPlanning bool) string {
	// Sanitize message
	message = strings.ReplaceAll(message, "\n", " ")
	message = strings.ReplaceAll(message, "\r", " ")
	message = strings.TrimSpace(message)

	// Trim message to the configured length with 3 dots
	maxMsgLen := TrimLength
	if len(message) > maxMsgLen && maxMsgLen > 3 {
		message = message[:maxMsgLen-3] + "..."
	}

	var output string
	if total == 0 {
		// During initial planning phase, show a fixed width bar
		width = 20 // Use a reasonable default width when total is unknown
		var bar string
		if isPlanning {
			planningText := "PLANNING"
			spacesBefore := (width - len(planningText)) / 2
			spacesAfter := width - len(planningText) - spacesBefore
			bar = strings.Repeat(" ", spacesBefore) + planningText + strings.Repeat(" ", spacesAfter)
		} else {
			bar = strings.Repeat("-", width)
		}
		output = fmt.Sprintf("[%s] %s", bar, message)
	} else {
		// Calculate progress bar width based on total steps
		barWidth := 20 // Default width
		if total > 0 {
			barWidth = total // Use total steps as width
		}

		// Calculate filled portion
		percent := float64(current) / float64(total)
		filledWidth := int(percent * float64(barWidth))
		if filledWidth < 0 {
			filledWidth = 0
		} else if filledWidth > barWidth {
			filledWidth = barWidth
		}

		var bar string
		if isPlanning {
			// Center "PLANNING" in the bar
			planningText := "PLANNING"
			spacesBefore := (barWidth - len(planningText)) / 2
			spacesAfter := barWidth - len(planningText) - spacesBefore
			bar = strings.Repeat(" ", spacesBefore) + planningText + strings.Repeat(" ", spacesAfter)
		} else {
			bar = strings.Repeat("=", filledWidth) + strings.Repeat(" ", barWidth-filledWidth)
		}
		output = fmt.Sprintf("(%02d)[%s](%02d) %s", current, bar, total, message)
	}

	return output
}
