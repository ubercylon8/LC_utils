package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"LC_utils/internal/api"
	"LC_utils/internal/auth"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var (
	// Global flags
	oid    string // Organization ID from flag
	apiKey string // API Key from flag
	action string
	fun    bool
	matrix bool
	hack   bool
	theme  string

	// Theme colors (package level)
	blue  = "\x1b[34m"
	cyan  = "\x1b[36m"
	white = "\x1b[37m"
	reset = "\x1b[0m"

	// List action flags
	output         string
	limit          int
	withTags       bool
	withIP         string
	hostnamePrefix string
	filterPlatform string
	filterHostname string
	filterTag      string
	onlineOnly     bool

	// Tag action flags
	sensorID   string
	addTags    []string
	removeTags []string

	// Task command flags
	taskCommand         string
	taskCommandList     string
	taskRandomDelay     bool
	taskPayloadName     string
	taskPayloadPath     string
	taskInvestigationID string
	taskReliable        bool
	taskContext         string

	// Upload payloads flags
	basePath  string
	outputFmt string
)

// Theme colors
type colorTheme struct {
	primary   string
	secondary string
	accent    string
	text      string
	reset     string
}

var themes = map[string]colorTheme{
	"matrix": {
		primary:   "\033[32m", // Green
		secondary: "\033[92m", // Bright Green
		accent:    "\033[32m", // Green
		text:      "\033[37m", // White
		reset:     "\033[0m",
	},
	"hacker": {
		primary:   "\033[31m", // Red
		secondary: "\033[91m", // Bright Red
		accent:    "\033[31m", // Red
		text:      "\033[37m", // White
		reset:     "\033[0m",
	},
	"cyberpunk": {
		primary:   "\033[35m", // Magenta
		secondary: "\033[36m", // Cyan
		accent:    "\033[93m", // Yellow
		text:      "\033[97m", // Bright White
		reset:     "\033[0m",
	},
	"retro": {
		primary:   "\033[33m", // Yellow
		secondary: "\033[93m", // Bright Yellow
		accent:    "\033[33m", // Yellow
		text:      "\033[37m", // White
		reset:     "\033[0m",
	},
}

func init() {
	// Force color output
	color.NoColor = false

	// Check environment variables first
	if envOID := os.Getenv("LC_ORG_ID"); envOID != "" {
		oid = envOID
	}
	if envAPIKey := os.Getenv("LC_API_KEY"); envAPIKey != "" {
		apiKey = envAPIKey
	}

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&oid, "oid", "o", oid, "LimaCharlie Organization ID (required)")
	rootCmd.PersistentFlags().StringVarP(&apiKey, "api-key", "k", apiKey, "LimaCharlie API Key (required)")
	rootCmd.Flags().BoolVar(&fun, "fun", false, "Just show the cool banner")
	rootCmd.Flags().BoolVar(&matrix, "matrix", false, "Show Matrix-style animation")
	rootCmd.Flags().BoolVar(&hack, "hack", false, "Show hacking animation")
	rootCmd.Flags().StringVar(&theme, "theme", "matrix", "Visual theme (matrix, hacker, cyberpunk, retro)")

	// Upload-payloads command flags
	uploadPayloadsCmd.Flags().StringVar(&basePath, "path", "", "Base path to search for executable files")
	uploadPayloadsCmd.Flags().StringVar(&outputFmt, "output", "json", "Output format (json or csv)")
	_ = uploadPayloadsCmd.MarkFlagRequired("path")

	// List command
	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List and filter sensors",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Use environment variables if available, otherwise use flags
			if oid == "" {
				return fmt.Errorf("organization ID is required (set via --oid flag or LC_ORG_ID environment variable)")
			}
			if apiKey == "" {
				return fmt.Errorf("API key is required (set via --api-key flag or LC_API_KEY environment variable)")
			}
			return nil
		},
		Run: runList,
	}

	// List command flags
	listCmd.Flags().StringVarP(&output, "output", "f", "text", "Output format (text/json/csv)")
	listCmd.Flags().IntVarP(&limit, "limit", "l", 0, "Limit the number of results")
	listCmd.Flags().BoolVarP(&withTags, "tags", "t", false, "Include sensor tags in output")
	listCmd.Flags().StringVarP(&withIP, "ip", "i", "", "Filter sensors by IP address")
	listCmd.Flags().StringVarP(&hostnamePrefix, "hostname", "n", "", "Filter sensors by hostname prefix")
	listCmd.Flags().StringVar(&filterHostname, "filter-hostname", "", "Filter by hostname (supports wildcards *)")
	listCmd.Flags().StringVar(&filterPlatform, "filter-platform", "", "Filter by platform (windows, macos)")
	listCmd.Flags().StringVar(&filterTag, "filter-tag", "", "Filter by tag")
	listCmd.Flags().BoolVar(&onlineOnly, "online", false, "Show only online sensors")

	// Tag command
	var tagCmd = &cobra.Command{
		Use:   "tag",
		Short: "Add or remove tags from a sensor",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if oid == "" {
				return fmt.Errorf("organization ID is required (set via --oid flag or LC_ORG_ID environment variable)")
			}
			if apiKey == "" {
				return fmt.Errorf("API key is required (set via --api-key flag or LC_API_KEY environment variable)")
			}
			if sensorID == "" {
				return fmt.Errorf("--sensor-id is required")
			}
			if len(addTags) == 0 && len(removeTags) == 0 {
				return fmt.Errorf("at least one of --add-tags or --remove-tags must be specified")
			}
			return nil
		},
		Run: runTag,
	}

	// Tag command flags
	tagCmd.Flags().StringVar(&sensorID, "sensor-id", "", "Sensor ID to tag (required)")
	tagCmd.Flags().StringSliceVar(&addTags, "add-tags", []string{}, "Tags to add (comma-separated)")
	tagCmd.Flags().StringSliceVar(&removeTags, "remove-tags", []string{}, "Tags to remove (comma-separated)")

	// Tag-multiple command
	var tagMultipleCmd = &cobra.Command{
		Use:   "tag-multiple",
		Short: "Add or remove tags from multiple sensors based on filters",
		Long: `Add or remove tags from all sensors that match the specified filters.
		
Example:
  # Tag all Windows sensors with hostname matching "*web*"
  lc-sensors tag-multiple -o ORG_ID -k API_KEY --filter-platform windows --filter-hostname "*web*" --add-tags web-server`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if oid == "" {
				return fmt.Errorf("organization ID is required (set via --oid flag or LC_ORG_ID environment variable)")
			}
			if apiKey == "" {
				return fmt.Errorf("API key is required (set via --api-key flag or LC_API_KEY environment variable)")
			}
			if len(addTags) == 0 && len(removeTags) == 0 {
				return fmt.Errorf("at least one of --add-tags or --remove-tags must be specified")
			}
			return nil
		},
		Run: runTagMultiple,
	}

	// Tag-multiple command flags
	tagMultipleCmd.Flags().StringVar(&filterHostname, "filter-hostname", "", "Filter by hostname (supports wildcards *)")
	tagMultipleCmd.Flags().StringVar(&filterPlatform, "filter-platform", "", "Filter by platform (windows, macos)")
	tagMultipleCmd.Flags().StringSliceVar(&addTags, "add-tags", []string{}, "Tags to add (comma-separated)")
	tagMultipleCmd.Flags().StringSliceVar(&removeTags, "remove-tags", []string{}, "Tags to remove (comma-separated)")

	// Task command
	var taskCmd = &cobra.Command{
		Use:   "task",
		Short: "Send tasks to sensors",
		Long: `Send tasks to LimaCharlie sensors.
		
Available Commands:
  put    Upload a file to sensors matching a hostname filter
  run    Execute a command on sensors matching a hostname filter`,
	}

	// Put command
	var putCmd = &cobra.Command{
		Use:   "put",
		Short: "Upload a file to sensors matching a hostname or tag filter",
		Long: `Upload a file to LimaCharlie sensors that match the specified hostname or tag filter.
		
Example:
  # Upload a file to all Windows sensors with hostname matching "web-*"
  lc-sensors task put -o ORG_ID -k API_KEY --filter-platform windows --filter-hostname "web-*" --payload-name file.txt --payload-path "/tmp/file.txt"`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if oid == "" {
				return fmt.Errorf("organization ID is required (set via --oid flag or LC_ORG_ID environment variable)")
			}
			if apiKey == "" {
				return fmt.Errorf("API key is required (set via --api-key flag or LC_API_KEY environment variable)")
			}
			if filterHostname == "" && filterTag == "" {
				return fmt.Errorf("either --filter-hostname or --filter-tag is required")
			}
			if taskCommandList == "" {
				if taskPayloadName == "" {
					return fmt.Errorf("--payload-name is required when not using --command-list")
				}
				if taskPayloadPath == "" {
					return fmt.Errorf("--payload-path is required when not using --command-list")
				}
			}
			return nil
		},
		Run: runPutTask,
	}

	// Put command flags
	putCmd.PersistentFlags().StringVar(&filterHostname, "filter-hostname", "", "Filter sensors by hostname (supports wildcards *)")
	putCmd.PersistentFlags().StringVar(&filterTag, "filter-tag", "", "Filter sensors by tag (supports wildcards *)")
	putCmd.PersistentFlags().StringVar(&filterPlatform, "filter-platform", "", "Filter by platform (windows, macos, linux)")
	putCmd.PersistentFlags().StringVar(&taskPayloadName, "payload-name", "", "Name of the payload file (required)")
	putCmd.PersistentFlags().StringVar(&taskPayloadPath, "payload-path", "", "Path on the sensor where to write the file (required)")
	putCmd.PersistentFlags().StringVar(&taskCommandList, "command-list", "", "Path to a file containing commands to execute (one per line)")
	putCmd.PersistentFlags().BoolVar(&taskRandomDelay, "random-delay", false, "Add random delay between commands (5-15 seconds)")
	putCmd.PersistentFlags().StringVar(&taskInvestigationID, "investigation-id", "", "Investigation ID to tag the task with")
	putCmd.PersistentFlags().BoolVar(&taskReliable, "reliable", false, "Use reliable tasking (will retry if sensor is offline)")
	putCmd.PersistentFlags().StringVar(&taskContext, "context", "", "Context value for reliable tasking (only used with --reliable)")

	// Run command
	var runCmd = &cobra.Command{
		Use:   "run",
		Short: "Execute a command on sensors matching a hostname filter",
		Long: `Execute a command on LimaCharlie sensors that match the specified hostname filter.
		
Example:
  # Run a command on all Windows sensors with hostname matching "web-*"
  lc-sensors task run -o ORG_ID -k API_KEY --filter-platform windows --filter-hostname "web-*" --command "whoami"`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if oid == "" {
				return fmt.Errorf("organization ID is required (set via --oid flag or LC_ORG_ID environment variable)")
			}
			if apiKey == "" {
				return fmt.Errorf("API key is required (set via --api-key flag or LC_API_KEY environment variable)")
			}
			if filterHostname == "" {
				return fmt.Errorf("--filter-hostname is required")
			}
			if taskCommand == "" && taskCommandList == "" {
				return fmt.Errorf("either --command or --command-list is required")
			}
			return nil
		},
		Run: runRunTask,
	}

	// Run command flags
	runCmd.PersistentFlags().StringVar(&filterHostname, "filter-hostname", "", "Filter sensors by hostname (supports wildcards *)")
	runCmd.PersistentFlags().StringVar(&filterPlatform, "filter-platform", "", "Filter by platform (windows, macos, linux)")
	runCmd.PersistentFlags().StringVar(&taskCommand, "command", "", "Command to execute (required if --command-list not specified)")
	runCmd.PersistentFlags().StringVar(&taskCommandList, "command-list", "", "Path to a file containing commands to execute (one per line)")
	runCmd.PersistentFlags().BoolVar(&taskRandomDelay, "random-delay", false, "Add random delay between commands (5-15 seconds)")
	runCmd.PersistentFlags().StringVar(&taskInvestigationID, "investigation-id", "", "Investigation ID to tag the task with")
	runCmd.PersistentFlags().BoolVar(&taskReliable, "reliable", false, "Use reliable tasking (will retry if sensor is offline)")
	runCmd.PersistentFlags().StringVar(&taskContext, "context", "", "Context value for reliable tasking (only used with --reliable)")

	// Add commands to task
	taskCmd.AddCommand(putCmd)
	taskCmd.AddCommand(runCmd)

	// Add all commands to root
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(tagCmd)
	rootCmd.AddCommand(tagMultipleCmd)
	rootCmd.AddCommand(taskCmd)
	rootCmd.AddCommand(uploadPayloadsCmd)
}

func getRandomMessage() string {
	messages := []string{
		"Remember: The best password is the one you can't remember 🔒",
		"I see you're trying to hack... ethically, of course! 🎯",
		"Warning: This tool may cause extreme levels of cybersecurity awareness 🚀",
		"Plot twist: The real malware was the friends we made along the way 🦠",
		"Scanning sensors... definitely not skynet... probably 🤖",
		"Fun fact: 'LC' stands for 'Literally Cybersecurity' (not really) 🛡️",
		"Loading 1337 h4x0r mode... just kidding, we're professionals here 👨‍💻",
		"Remember to thank your local SOC analyst today! 🏆",
		"Roses are red, violets are blue, this tool is secure, and your network should be too 🌹",
		"Keep calm and incident response on 🚨",
		"chmod 777 is like giving everyone a key to your house... don't do it! 🏠",
		"I'm not saying I'm Batman, but have you ever seen me and Batman's SIEM in the same room? 🦇",
		"Error 404: Malware not found (that's a good thing) 🎉",
		"Your network is like an onion, it has layers... and might make you cry ��",
		"Firewall rule #1: Deny all, allow none, then panic when nothing works 🧱",
		"I would tell you a UDP joke, but you might not get it 📡",
		"Why do Java developers wear glasses? Because they don't C# 👓",
		"There are 10 types of people in the world: those who understand binary, and those who don't 🤓",
		"Keep your friends close, and your private keys closer 🔑",
		"I'm not lazy, I'm just running in low-power mode 🔋",
		"Trust me, I'm a security scanner 🔍",
		"My password is so strong, even I can't remember it 💪",
		"In case of fire: git commit, git push, exit building 🔥",
		"I'm not a robot... but that's exactly what a robot would say 🤖",
		"Knock knock! Who's there? Very long pause... Java ��",
		"Documentation? You mean the code comments I'll write someday? 📝",
		"May the brute force NOT be with you 🚫",
		"I put the 'Sec' in DevSecOps... and the 'Ops' in 'Oops' 😅",
		"Trying to hack this? Good luck, I'm behind 7 proxies 🎭",
		"Alert: High CPU usage detected... oh wait, that's just Chrome 🌐",
		"Quantum computing called, it wants its encryption back 📡",
		"My code is so secure, it won't even let me in 🚪",
		"Eat, Sleep, Patch, Repeat 🔄",
		"I don't always test my code, but when I do, I do it in production 🎲",
		"Why did the security engineer bring a ladder to work? To climb over the firewall 🪜",
		"Scanning ports like it's 1999 📊",
		"Zero-day? More like zero-chill 😎",
		"This message is ROT13 encrypted... twice! 🔐",
		"I would explain DNS to you, but it might take 24-48 hours to propagate 🌍",
		"Error 418: I'm a teapot (RFC 2324) ☕",
	}
	rand.Seed(time.Now().UnixNano())
	return messages[rand.Intn(len(messages))]
}

func printBanner() string {
	var sb strings.Builder

	// ANSI color codes
	blue := "\x1b[34m"
	cyan := "\x1b[36m"
	white := "\x1b[37m"
	reset := "\x1b[0m"

	// Create a sophisticated cybersecurity-themed banner
	banner := `
    ██╗      ██████╗     ██████╗ ██████╗ ███╗   ███╗███╗   ███╗ █████╗ ███╗   ██╗██████╗ ██████╗ ██████╗ 
    ██║     ██╔════╝    ██╔════╝██╔═████╗████╗ ████║████╗ ████║██╔══██╗████╗  ██║██╔══██╗╚════██╗██╔══██╗
    ██║     ██║         ██║     ██║██╔██║██╔████╔██║██╔████╔██║███████║██╔██╗ ██║██║  ██║ █████╔╝██████╔╝
    ██║     ██║         ██║     ████╔╝██║██║╚██╔╝██║██║╚██╔╝██║██╔══██║██║╚██╗██║██║  ██║ ╚═══██╗██╔══██╗
    ███████╗╚██████╗    ╚██████╗╚██████╔╝██║ ╚═╝ ██║██║ ╚═╝ ██║██║  ██║██║ ╚████║██████╔╝██████╔╝██║  ██║
    ╚══════╝ ╚═════╝     ╚═════╝ ╚═════╝ ╚═╝     ╚═╝╚═╝     ╚═╝╚═╝  ╚═╝╚═╝  ╚═══╝╚═════╝ ╚═════╝ ╚═╝  ╚═╝`

	sb.WriteString("\n") // Initial spacing

	// Print the banner with gradient effect
	lines := strings.Split(banner, "\n")
	for i, line := range lines {
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}
		// Create a gradient effect
		switch i % 3 {
		case 0:
			sb.WriteString(fmt.Sprintf("%s%s%s\n", blue, line, reset))
		case 1:
			sb.WriteString(fmt.Sprintf("%s%s%s\n", cyan, line, reset))
		case 2:
			sb.WriteString(fmt.Sprintf("%s%s%s\n", white, line, reset))
		}
	}

	// Add a cybersecurity-themed subtitle with colors
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("%s╔════ %s", cyan, reset))
	sb.WriteString(fmt.Sprintf("%sSecure Command & Control%s", white, reset))
	sb.WriteString(fmt.Sprintf("%s ════╗%s", cyan, reset))
	sb.WriteString("\n\n")

	// Add version and status with colors
	sb.WriteString(fmt.Sprintf("%s[%s", blue, reset))
	sb.WriteString(fmt.Sprintf("%sv1.0.0%s", white, reset))
	sb.WriteString(fmt.Sprintf("%s] %s", blue, reset))
	sb.WriteString(fmt.Sprintf("%sSystem Status: %s", cyan, reset))
	sb.WriteString(fmt.Sprintf("%sOPERATIONAL%s", white, reset))
	sb.WriteString("\n\n")

	// Add random message with a cool prefix
	sb.WriteString(fmt.Sprintf("%s⚡ %s", cyan, reset))
	sb.WriteString(fmt.Sprintf("%s%s%s", white, getRandomMessage(), reset))
	sb.WriteString("\n\n")

	// Add a decorative footer
	sb.WriteString(fmt.Sprintf("%s═══════════════════════════════════════════════════%s", cyan, reset))
	sb.WriteString("\n\n")

	return sb.String()
}

func printMatrixEffect() {
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789@#$%&*"
	green := "\033[32m"
	brightGreen := "\033[92m"
	reset := "\033[0m"

	// Clear screen
	fmt.Print("\033[H\033[2J")

	// Create channels for each column
	columns := 80
	drops := make([]int, columns)
	for i := range drops {
		drops[i] = -1 * (rand.Intn(20) + 1) // Start above screen at random heights
	}

	// Run animation for a few seconds
	start := time.Now()
	for time.Since(start) < 3*time.Second {
		// Clear screen and move to top
		fmt.Print("\033[H")

		// Update and draw each column
		for i := range drops {
			// Move drop down
			drops[i]++
			if drops[i] > 20 { // Screen height
				drops[i] = 0
			}

			// Draw the column
			for j := 0; j < 20; j++ { // Screen height
				if j == drops[i] {
					// Head of the drop
					fmt.Printf("%s%c%s", brightGreen, chars[rand.Intn(len(chars))], reset)
				} else if j < drops[i] && j > drops[i]-5 {
					// Tail of the drop
					fmt.Printf("%s%c%s", green, chars[rand.Intn(len(chars))], reset)
				} else {
					fmt.Print(" ")
				}
			}
			fmt.Print("\n")
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func getCurrentTheme() colorTheme {
	if t, ok := themes[theme]; ok {
		return t
	}
	return themes["matrix"] // Default theme
}

func printHackingAnimation() {
	t := getCurrentTheme()
	steps := []struct {
		message  string
		progress int
	}{
		{"Initializing cyber protocols", 10},
		{"Bypassing mainframe security", 20},
		{"Decrypting quantum algorithms", 35},
		{"Compiling zero-day exploits", 48},
		{"Routing through proxy chain", 60},
		{"Deploying cyber countermeasures", 75},
		{"Establishing secure connection", 85},
		{"Finalizing system access", 95},
		{"Access granted", 100},
	}

	// Clear screen
	fmt.Print("\033[H\033[2J")

	for _, step := range steps {
		// Calculate progress bar
		width := 50
		filled := width * step.progress / 100
		bar := strings.Repeat("▓", filled) + strings.Repeat("░", width-filled)

		// Print progress with theme colors
		fmt.Printf("\r%s[%s] %d%%%s %s%s%s",
			t.primary, bar, step.progress, t.reset,
			t.accent, step.message, t.reset)

		// Add some "hacking" details below the progress bar
		details := []string{
			"0x" + fmt.Sprintf("%X", rand.Int31()),
			fmt.Sprintf("PORT_%d_SCAN_COMPLETE", rand.Intn(65535)),
			fmt.Sprintf("THREAD_%d_INITIALIZED", rand.Intn(100)),
			fmt.Sprintf("BUFFER_%X_OVERFLOW_CHECKED", rand.Intn(0xFFFF)),
		}

		// Print random technical details
		fmt.Println()
		for i := 0; i < 3; i++ {
			detail := details[rand.Intn(len(details))]
			fmt.Printf("%s%s%s\n", t.secondary, detail, t.reset)
		}
		fmt.Println()

		// Random delay between steps
		time.Sleep(time.Duration(500+rand.Intn(1000)) * time.Millisecond)

		// Clear lines for next update (progress + 3 details + empty line = 5 lines)
		fmt.Print("\033[5A")
	}

	// Final message
	fmt.Println("\n" + t.accent + "SYSTEM COMPROMISED - Access Level: ROOT" + t.reset + "\n")
	time.Sleep(1 * time.Second)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		color.Red("Error: %v", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "lc-sensors",
	Short: "LimaCharlie Sensor Management Tool",
	Long:  `A CLI tool for managing LimaCharlie sensors and related functionality.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Environment variables are already handled in init()
		// Apply theme to all color output
		if theme != "" {
			t := getCurrentTheme()
			color.NoColor = false
			blue = t.primary
			cyan = t.secondary
			white = t.text
			reset = t.reset
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		// If hack flag is set, show the hacking animation and exit
		if hack {
			printHackingAnimation()
			fmt.Print(printBanner())
			return
		}

		// If matrix flag is set, show the animation and exit
		if matrix {
			printMatrixEffect()
			fmt.Print(printBanner())
			return
		}

		// If fun flag is set or no arguments provided, just show the banner and help
		if fun || len(args) == 0 {
			fmt.Print(printBanner())
			if !fun {
				cmd.Help()
			}
			return
		}
	},
}

var uploadPayloadsCmd = &cobra.Command{
	Use:   "upload-payloads",
	Short: "Upload multiple executable files as payloads",
	Long: `Upload multiple executable files as payloads to LimaCharlie.
This command will recursively search for .exe files in the specified directory
and upload them as payloads to your LimaCharlie organization.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if oid == "" || apiKey == "" {
			return fmt.Errorf("organization ID and API key are required")
		}

		// Find all executable files
		files, err := api.FindExecutableFiles(basePath)
		if err != nil {
			return fmt.Errorf("error finding executable files: %w", err)
		}

		if len(files) == 0 {
			color.Yellow("No executable files found in %s\n", basePath)
			return nil
		}

		// Process each file
		results := make(map[string]string)
		for _, file := range files {
			relPath, _ := filepath.Rel(basePath, file)
			fmt.Printf("Processing %s... ", relPath)

			err := api.UploadPayload(oid, apiKey, file)
			if err != nil {
				results[relPath] = fmt.Sprintf("Error: %v", err)
				color.Red("Failed")
			} else {
				results[relPath] = "Success"
				color.Green("Success")
			}
		}

		// Output results
		switch outputFmt {
		case "json":
			jsonData, err := json.MarshalIndent(results, "", "  ")
			if err != nil {
				return fmt.Errorf("error formatting JSON output: %w", err)
			}
			fmt.Println(string(jsonData))
		case "csv":
			fmt.Println("File,Status")
			for file, status := range results {
				fmt.Printf("%s,%s\n", file, status)
			}
		default:
			return fmt.Errorf("unsupported output format: %s", outputFmt)
		}

		return nil
	},
}

func runList(cmd *cobra.Command, args []string) {
	// Print banner
	fmt.Print(printBanner())

	// Validate required flags
	if oid == "" || apiKey == "" {
		color.Red("Error: --oid and --api-key are required")
		os.Exit(1)
	}

	// Initialize credentials
	creds := &auth.Credentials{
		OID:    oid,
		APIKey: apiKey,
	}

	// Validate credentials
	if err := creds.ValidateCredentials(); err != nil {
		color.Red("Error: %v", err)
		os.Exit(1)
	}

	color.Blue("Authenticating with LimaCharlie...")

	// Prepare listing options
	opts := &api.ListOptions{
		Limit:              limit,
		WithTags:           withTags || filterTag != "", // Always fetch tags if filtering by tag
		WithIP:             withIP,
		WithHostnamePrefix: hostnamePrefix,
		OnlyOnline:         onlineOnly,
	}

	// List sensors
	color.Blue("Retrieving sensors...")
	sensors, err := api.ListSensors(creds, opts)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "401") {
			color.Red("Authentication failed. Please check your API key and organization ID.")
		} else {
			color.Red("Failed to retrieve sensors: %v", err)
		}
		os.Exit(1)
	}

	color.Green("Successfully authenticated with LimaCharlie!")

	// Apply all filters
	if filterHostname != "" || filterPlatform != "" || filterTag != "" {
		sensors = filterSensors(sensors, nil)
	}

	// Output results based on format
	outputResults(sensors)
}

func runTag(cmd *cobra.Command, args []string) {
	// Print banner
	fmt.Print(printBanner())

	// Validate required flags
	if oid == "" || apiKey == "" {
		color.Red("Error: --oid and --api-key are required")
		os.Exit(1)
	}

	if len(addTags) == 0 && len(removeTags) == 0 {
		color.Red("Error: at least one of --add-tags or --remove-tags must be specified")
		os.Exit(1)
	}

	// Initialize credentials
	creds := &auth.Credentials{
		OID:    oid,
		APIKey: apiKey,
	}

	// Tag a single sensor
	if err := api.TagSensor(creds, sensorID, api.TagSensorRequest{
		AddTags:    addTags,
		RemoveTags: removeTags,
	}); err != nil {
		color.Red("Failed to tag sensor: %w", err)
		os.Exit(1)
	}

	// Success message
	color.Green("\nSuccessfully updated tags for sensor %s", sensorID)
	if len(addTags) > 0 {
		color.Blue("Added tags: %v", addTags)
	}
	if len(removeTags) > 0 {
		color.Blue("Removed tags: %v", removeTags)
	}
}

func runTagMultiple(cmd *cobra.Command, args []string) {
	// Print banner
	fmt.Print(printBanner())

	// Validate required flags
	if oid == "" || apiKey == "" {
		color.Red("Error: --oid and --api-key are required")
		os.Exit(1)
	}

	if len(addTags) == 0 && len(removeTags) == 0 {
		color.Red("Error: at least one of --add-tags or --remove-tags must be specified")
		os.Exit(1)
	}

	// Initialize credentials
	creds := &auth.Credentials{
		OID:    oid,
		APIKey: apiKey,
	}

	// List all sensors with their tags
	opts := &api.ListOptions{
		WithTags: true, // We need tags for proper filtering
	}

	// List sensors
	color.Blue("Retrieving sensors...")
	sensors, err := api.ListSensors(creds, opts)
	if err != nil {
		color.Red("Failed to retrieve sensors: %v", err)
		os.Exit(1)
	}

	// Filter sensors based on hostname and platform
	filtered := filterSensors(sensors, nil)

	if len(filtered) == 0 {
		color.Yellow("No sensors match the specified filters")
		os.Exit(0)
	}

	// Confirm with user
	color.Yellow("\nFound %d sensors matching filters:", len(filtered))
	for _, sensor := range filtered {
		fmt.Printf("- %s (%s) [%s]\n", sensor.Hostname, sensor.SID, sensor.GetPlatformString())
	}

	fmt.Print("\nDo you want to proceed with tagging these sensors? [y/N] ")
	var response string
	fmt.Scanln(&response)
	if strings.ToLower(response) != "y" {
		color.Yellow("Operation cancelled")
		os.Exit(0)
	}

	// Tag multiple sensors
	// Tag each sensor
	color.Blue("\nUpdating sensor tags...")
	for _, sensor := range filtered {
		if err := api.TagSensor(creds, sensor.SID, api.TagSensorRequest{
			AddTags:    addTags,
			RemoveTags: removeTags,
		}); err != nil {
			color.Red("Failed to tag sensor %s: %v", sensor.SID, err)
			os.Exit(1)
		}
		color.Green("Successfully tagged sensor %s", sensor.SID)
	}

	// Print summary
	fmt.Println()
	if len(addTags) > 0 {
		color.Green("Successfully tagged %d sensors with added tags: %v", len(filtered), addTags)
	}
	if len(removeTags) > 0 {
		color.Green("Successfully tagged %d sensors with removed tags: %v", len(filtered), removeTags)
	}
}

func outputResults(sensors []api.Sensor) {
	switch output {
	case "json":
		outputJSON(sensors)
	case "csv":
		outputCSV(sensors)
	default:
		outputText(sensors)
	}
}

func filterSensors(sensors []api.Sensor, onlineStatuses *api.OnlineStatusResponse) []api.Sensor {
	var filtered []api.Sensor

	for _, sensor := range sensors {
		include := true

		// Filter by hostname if specified
		if filterHostname != "" {
			pattern := strings.ReplaceAll(filterHostname, "*", ".*")
			matched, err := regexp.MatchString(pattern, sensor.Hostname)
			if err != nil || !matched {
				include = false
				continue
			}
		}

		// Filter by platform if specified
		if filterPlatform != "" {
			platformStr := strings.ToLower(sensor.GetPlatformString())
			if !strings.EqualFold(platformStr, filterPlatform) {
				include = false
				continue
			}
		}

		// Filter by tag if specified
		if filterTag != "" {
			pattern := strings.ReplaceAll(filterTag, "*", ".*")
			tagFound := false
			for _, tag := range sensor.Tags {
				matched, err := regexp.MatchString(pattern, tag)
				if err == nil && matched {
					tagFound = true
					break
				}
			}
			if !tagFound {
				include = false
				continue
			}
		}

		// Filter by online status if specified
		if onlineOnly && !sensor.IsOnline {
			include = false
			continue
		}

		if include {
			filtered = append(filtered, sensor)
		}
	}

	return filtered
}

func outputJSON(sensors []api.Sensor) {
	jsonOutput, err := json.MarshalIndent(sensors, "", "  ")
	if err != nil {
		color.Red("Failed to format JSON: %v", err)
		os.Exit(1)
	}
	fmt.Println(string(jsonOutput))
}

func outputCSV(sensors []api.Sensor) {
	// Print sensor details in CSV format
	w := csv.NewWriter(os.Stdout)
	w.Write([]string{"SID", "Hostname", "Platform", "Architecture", "Last Seen", "Enrollment Time", "External IP", "Internal IP", "Online", "Tags"})

	for _, sensor := range sensors {
		// Format tags
		tagsStr := strings.Join(sensor.Tags, ", ")

		w.Write([]string{
			sensor.SID,
			sensor.Hostname,
			sensor.GetPlatformString(),
			sensor.GetArchitectureString(),
			sensor.GetLastSeenString(),
			sensor.GetEnrollmentTimeString(),
			sensor.ExternalIP,
			sensor.InternalIP,
			fmt.Sprintf("%v", sensor.IsOnline),
			tagsStr,
		})
	}

	w.Flush()
}

func outputText(sensors []api.Sensor) {
	color.Green("\nFound %d sensors:", len(sensors))
	fmt.Println("\n" + color.New(color.FgHiBlack).Sprint("─────────────────────────────────────────"))

	// Print sensor details in table format
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"SID", "Hostname", "Platform", "Status", "IP", "Last Seen", "Tags"})
	table.SetBorder(false)

	for _, sensor := range sensors {
		status := "OFFLINE"
		if sensor.IsOnline {
			status = "ONLINE"
		}

		// Format tags
		tagsStr := strings.Join(sensor.Tags, ", ")

		table.Append([]string{
			sensor.SID,
			sensor.Hostname,
			sensor.GetPlatformString(),
			status,
			sensor.ExternalIP,
			sensor.GetLastSeenString(),
			tagsStr,
		})
	}

	table.Render()
}

func outputSensorText(sensor api.Sensor) {
	fmt.Printf("\nSensor Details:\n")
	fmt.Printf("SID: %s\n", sensor.SID)
	fmt.Printf("Hostname: %s\n", sensor.Hostname)
	fmt.Printf("Platform: %s\n", sensor.GetPlatformString())
	fmt.Printf("Architecture: %s\n", sensor.GetArchitectureString())
	fmt.Printf("Last Seen: %s\n", sensor.GetLastSeenString())
	fmt.Printf("Enrollment Time: %s\n", sensor.GetEnrollmentTimeString())
	fmt.Printf("External IP: %s\n", sensor.ExternalIP)
	fmt.Printf("Internal IP: %s\n", sensor.InternalIP)
	fmt.Printf("Version: %s\n", sensor.Version)
	fmt.Printf("Online: %v\n", sensor.IsOnline)

	// Format and display tags
	if len(sensor.Tags) > 0 {
		fmt.Printf("Tags: %s\n", strings.Join(sensor.Tags, ", "))
	} else {
		fmt.Printf("Tags: None\n")
	}
}

func runRunTask(cmd *cobra.Command, args []string) {
	// Print banner
	fmt.Print(printBanner())

	// Initialize credentials
	creds := &auth.Credentials{
		OID:    oid,
		APIKey: apiKey,
	}

	// List all sensors
	color.Blue("Retrieving sensors...")
	opts := &api.ListOptions{
		WithTags: false, // We don't need tags for hostname filtering
	}

	sensors, err := api.ListSensors(creds, opts)
	if err != nil {
		color.Red("Failed to retrieve sensors: %v", err)
		os.Exit(1)
	}

	// Filter sensors based on hostname
	filtered := filterSensors(sensors, nil)

	if len(filtered) == 0 {
		color.Yellow("No sensors match the hostname filter: %s", filterHostname)
		os.Exit(0)
	}

	// Confirm with user
	color.Yellow("\nFound %d sensors matching hostname filter '%s':", len(filtered), filterHostname)
	for _, sensor := range filtered {
		fmt.Printf("- %s (%s)\n", sensor.Hostname, sensor.SID)
	}

	fmt.Print("\nDo you want to proceed with running the command on these sensors? [y/N] ")
	var response string
	fmt.Scanln(&response)
	if strings.ToLower(response) != "y" {
		color.Yellow("Operation cancelled")
		os.Exit(0)
	}

	// Get commands to execute
	var commands []string
	if taskCommandList != "" {
		var err error
		commands, err = readCommandsFromFile(taskCommandList)
		if err != nil {
			color.Red("Failed to read command list: %v", err)
			os.Exit(1)
		}
	} else {
		commands = []string{taskCommand}
	}

	// Run commands on each sensor
	color.Blue("\nExecuting commands on sensors...")
	var successCount, failCount int
	for _, sensor := range filtered {
		for i, command := range commands {
			if i > 0 {
				addRandomDelay()
			}

			if taskReliable {
				// Use reliable tasking
				if err := api.CreateReliableTask(creds, sensor.SID, command, taskContext); err != nil {
					color.Red("Failed to send reliable task to sensor %s (%s): %v", sensor.Hostname, sensor.SID, err)
					failCount++
				} else {
					color.Green("Successfully queued reliable task for sensor %s (%s): %s", sensor.Hostname, sensor.SID, command)
					successCount++
				}
			} else {
				// Use regular tasking
				if _, err := api.RunCommand(creds, sensor.SID, command, taskInvestigationID); err != nil {
					color.Red("Failed to run command on sensor %s (%s): %v", sensor.Hostname, sensor.SID, err)
					failCount++
				} else {
					color.Green("Successfully sent command to sensor %s (%s): %s", sensor.Hostname, sensor.SID, command)
					successCount++
				}
			}
		}
	}

	// Print summary
	fmt.Println()
	if successCount > 0 {
		if taskReliable {
			color.Green("Successfully queued reliable command for %d sensors", successCount)
		} else {
			color.Green("Successfully sent command to %d sensors", successCount)
		}
	}
	if failCount > 0 {
		if taskReliable {
			color.Red("Failed to queue reliable command for %d sensors", failCount)
		} else {
			color.Red("Failed to send command to %d sensors", failCount)
		}
	}
	if successCount > 0 {
		color.Yellow("\nNote: Command output is not available through the API. Check the LimaCharlie web interface for results.")
	}
}

func runPutTask(cmd *cobra.Command, args []string) {
	// Print banner
	fmt.Print(printBanner())

	// Initialize credentials
	creds := &auth.Credentials{
		OID:    oid,
		APIKey: apiKey,
	}

	// List all sensors
	color.Blue("Retrieving sensors...")
	opts := &api.ListOptions{
		WithTags: filterTag != "", // Only fetch tags if filtering by tag
	}

	sensors, err := api.ListSensors(creds, opts)
	if err != nil {
		color.Red("Failed to retrieve sensors: %v", err)
		os.Exit(1)
	}

	// Filter sensors based on hostname and/or tag
	filtered := filterSensors(sensors, nil)

	if len(filtered) == 0 {
		if filterHostname != "" {
			color.Yellow("No sensors match the hostname filter: %s", filterHostname)
		}
		if filterTag != "" {
			color.Yellow("No sensors match the tag filter: %s", filterTag)
		}
		os.Exit(0)
	}

	// Confirm with user
	if filterHostname != "" && filterTag != "" {
		color.Yellow("\nFound %d sensors matching hostname filter '%s' and tag filter '%s':", len(filtered), filterHostname, filterTag)
	} else if filterHostname != "" {
		color.Yellow("\nFound %d sensors matching hostname filter '%s':", len(filtered), filterHostname)
	} else {
		color.Yellow("\nFound %d sensors matching tag filter '%s':", len(filtered), filterTag)
	}

	for _, sensor := range filtered {
		if len(sensor.Tags) > 0 {
			fmt.Printf("- %s (%s) [Tags: %v]\n", sensor.Hostname, sensor.SID, sensor.Tags)
		} else {
			fmt.Printf("- %s (%s)\n", sensor.Hostname, sensor.SID)
		}
	}

	fmt.Print("\nDo you want to proceed with uploading the file to these sensors? [y/N] ")
	var response string
	fmt.Scanln(&response)
	if strings.ToLower(response) != "y" {
		color.Yellow("Operation cancelled")
		os.Exit(0)
	}

	// Get commands to execute
	var commands []string
	if taskCommandList != "" {
		var err error
		commands, err = readCommandsFromFile(taskCommandList)
		if err != nil {
			color.Red("Failed to read command list: %v", err)
			os.Exit(1)
		}
	} else {
		task := fmt.Sprintf("put --payload-name %s --payload-path '%s'", taskPayloadName, taskPayloadPath)
		commands = []string{task}
	}

	// Run commands on each sensor
	color.Blue("\nUploading files to sensors...")
	var successCount, failCount int
	for _, sensor := range filtered {
		for i, command := range commands {
			if i > 0 {
				addRandomDelay()
			}

			if taskReliable {
				// Prepare reliable tasking request data
				data := map[string]interface{}{
					"task": command,
					"sid":  sensor.SID,
					"ttl":  3600, // 1 hour TTL
				}

				// Add context if provided
				if taskContext != "" {
					data["context"] = taskContext
				} else if taskInvestigationID != "" {
					data["context"] = taskInvestigationID
				}

				jsonData, err := json.Marshal(data)
				if err != nil {
					color.Red("Failed to prepare reliable task for sensor %s (%s): %v", sensor.Hostname, sensor.SID, err)
					failCount++
					continue
				}

				// Send reliable task request
				if err := api.CreateExtensionRequest(creds, "ext-reliable-tasking", "task", string(jsonData)); err != nil {
					color.Red("Failed to send reliable task to sensor %s (%s): %v", sensor.Hostname, sensor.SID, err)
					failCount++
				} else {
					color.Green("Successfully queued reliable task for sensor %s (%s): %s", sensor.Hostname, sensor.SID, command)
					successCount++
				}
			} else {
				// Use regular tasking
				if _, err := api.TaskSensor(creds, sensor.SID, []string{command}, taskInvestigationID); err != nil {
					color.Red("Failed to upload file to sensor %s (%s): %v", sensor.Hostname, sensor.SID, err)
					failCount++
				} else {
					color.Green("Successfully sent upload command to sensor %s (%s): %s", sensor.Hostname, sensor.SID, command)
					successCount++
				}
			}
		}
	}

	// Print summary
	fmt.Println()
	if successCount > 0 {
		if taskReliable {
			color.Green("Successfully queued reliable upload task for %d sensors", successCount)
		} else {
			color.Green("Successfully sent upload command to %d sensors", successCount)
		}
	}
	if failCount > 0 {
		if taskReliable {
			color.Red("Failed to queue reliable upload task for %d sensors", failCount)
		} else {
			color.Red("Failed to send upload command to %d sensors", failCount)
		}
	}
	if successCount > 0 {
		color.Yellow("\nNote: Upload status is not available through the API. Check the LimaCharlie web interface for results.")
	}
}

// Add helper function to read commands from file
func readCommandsFromFile(filePath string) ([]string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading command file: %w", err)
	}

	// Split content by newlines and filter empty lines
	var commands []string
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			commands = append(commands, line)
		}
	}

	if len(commands) == 0 {
		return nil, fmt.Errorf("command file is empty")
	}

	return commands, nil
}

// Add random delay function
func addRandomDelay() {
	if taskRandomDelay {
		delay := time.Duration(5+rand.Intn(11)) * time.Second // Random delay between 5-15 seconds
		color.Yellow("Waiting %v before next command...", delay)
		time.Sleep(delay)
	}
}
