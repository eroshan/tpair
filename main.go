package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

const (
	charSet      = "sadjklewcmpgh"
	defaultPairs = 10
)

// Global stats variable for signal handling
var globalStats Stats

type Stats struct {
	Total     int
	Correct   int
	Errors    int
	TotalTime time.Duration
}

func (s *Stats) Accuracy() float64 {
	if s.Total == 0 {
		return 0
	}
	return float64(s.Correct) / float64(s.Total) * 100
}

func (s *Stats) WPM() float64 {
	if s.TotalTime.Minutes() == 0 {
		return 0
	}
	// Assuming average word length of 5 characters for WPM calculation
	words := float64(s.Correct*2) / 5
	return words / s.TotalTime.Minutes()
}

func generateRandomPair() string {
	first := charSet[rand.Intn(len(charSet))]
	second := charSet[rand.Intn(len(charSet))]
	return string(first) + string(second)
}

// Terminal control functions for raw input
func enableRawMode() *syscall.Termios {
	var oldState syscall.Termios
	fd := int(os.Stdin.Fd())

	// Get current terminal state
	syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), syscall.TIOCGETA, uintptr(unsafe.Pointer(&oldState)))

	// Create new state with raw mode
	newState := oldState
	newState.Lflag &^= syscall.ECHO | syscall.ICANON
	newState.Cc[syscall.VMIN] = 1
	newState.Cc[syscall.VTIME] = 0

	// Set raw mode
	syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), syscall.TIOCSETA, uintptr(unsafe.Pointer(&newState)))

	return &oldState
}

func disableRawMode(oldState *syscall.Termios) {
	fd := int(os.Stdin.Fd())
	syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), syscall.TIOCSETA, uintptr(unsafe.Pointer(oldState)))
}

func readChar() (byte, error) {
	var buf [1]byte
	n, err := os.Stdin.Read(buf[:])
	if n == 0 || err != nil {
		return 0, err
	}
	return buf[0], nil
}

func printWelcome() {
	fmt.Println("╔═══════════════════════════════════════╗")
	fmt.Println("║         Double Typer Practice         ║")
	fmt.Println("║   Practice typing character pairs!   ║")
	fmt.Println("╚═══════════════════════════════════════╝")
	fmt.Printf("Character set: %s\n", charSet)
	fmt.Println("Type the character pairs as they appear.")
	fmt.Println("Press Ctrl+C to exit and see statistics.")
	fmt.Println("No need to press Enter - just type the pairs!")
	fmt.Println()
}

func printStats(stats Stats) {
	fmt.Println("\n" + strings.Repeat("=", 40))
	fmt.Println("           SESSION STATISTICS")
	fmt.Println(strings.Repeat("=", 40))
	fmt.Printf("Total pairs:     %d\n", stats.Total)
	fmt.Printf("Correct:         %d\n", stats.Correct)
	fmt.Printf("Errors:          %d\n", stats.Errors)
	fmt.Printf("Accuracy:        %.1f%%\n", stats.Accuracy())
	fmt.Printf("Average WPM:     %.1f\n", stats.WPM())
	fmt.Printf("Total time:      %.1fs\n", stats.TotalTime.Seconds())
	fmt.Println(strings.Repeat("=", 40))
}

func getAccuracyEmoji(accuracy float64) string {
	switch {
	case accuracy >= 95:
		return "🎯"
	case accuracy >= 85:
		return "👍"
	case accuracy >= 75:
		return "👌"
	case accuracy >= 65:
		return "😊"
	default:
		return "💪"
	}
}

func runTypingSession() {
	// Enable raw mode for character-by-character input
	oldState := enableRawMode()
	defer disableRawMode(oldState)

	for {
		pair := generateRandomPair()
		// Highlight the characters to type with background color, bold, and uppercase
		fmt.Printf("Type: \033[1;37;44m %s \033[0m -> ", strings.ToUpper(pair))

		start := time.Now()
		var input string

		// Read exactly 2 characters
		for i := 0; i < 2; i++ {
			char, err := readChar()
			if err != nil {
				fmt.Println("\nError reading input:", err)
				return
			}

			input += string(char)
			fmt.Print(string(char)) // Echo the character
		}

		elapsed := time.Since(start)

		globalStats.Total++
		globalStats.TotalTime += elapsed

		if input == pair {
			globalStats.Correct++
			fmt.Print(" ✓ Correct! ")
		} else {
			globalStats.Errors++
			fmt.Printf(" ✗ Wrong! (Expected: %s) ", pair)
		}

		// Show live stats every 5 pairs
		if globalStats.Total%5 == 0 {
			fmt.Printf("| Accuracy: %.1f%% | WPM: %.1f %s\n",
				globalStats.Accuracy(), globalStats.WPM(), getAccuracyEmoji(globalStats.Accuracy()))
		} else {
			fmt.Println()
		}
	}
}

func printUsage() {
	fmt.Println("Usage: doubletyper [options]")
	fmt.Println("\nOptions:")
	fmt.Println("  help, -h, --help    Show this help message")
	fmt.Println("  stats               Show character pair statistics")
	fmt.Println("\nDuring practice:")
	fmt.Println("  Type character pairs directly (no Enter needed)")
	fmt.Println("  Press Ctrl+C to exit and see statistics")
}

func showCharacterStats() {
	fmt.Println("Character Pair Possibilities:")
	fmt.Println(strings.Repeat("-", 30))

	count := 0
	for i := 0; i < len(charSet); i++ {
		for j := 0; j < len(charSet); j++ {
			fmt.Printf("%c%c ", charSet[i], charSet[j])
			count++
			if count%10 == 0 {
				fmt.Println()
			}
		}
	}
	if count%10 != 0 {
		fmt.Println()
	}
	fmt.Printf("\nTotal possible pairs: %d\n", len(charSet)*len(charSet))
	fmt.Printf("Character set: %s (%d characters)\n", charSet, len(charSet))
}

func setupSignalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("\n^C")

		if globalStats.Total > 0 {
			fmt.Println("\n" + strings.Repeat("=", 40))
			fmt.Println("           SESSION STATISTICS")
			fmt.Println(strings.Repeat("=", 40))
			fmt.Printf("Total pairs:     %d\n", globalStats.Total)
			fmt.Printf("Correct:         %d\n", globalStats.Correct)
			fmt.Printf("Errors:          %d\n", globalStats.Errors)
			fmt.Printf("Accuracy:        %.1f%%\n", globalStats.Accuracy())
			fmt.Printf("Average WPM:     %.1f\n", globalStats.WPM())
			fmt.Printf("Total time:      %.1fs\n", globalStats.TotalTime.Seconds())
			fmt.Println(strings.Repeat("=", 40))

			// Provide encouragement based on performance
			accuracy := globalStats.Accuracy()
			switch {
			case accuracy >= 95:
				fmt.Println("🎉 Outstanding! You're a typing master!")
			case accuracy >= 85:
				fmt.Println("🌟 Excellent work! Keep it up!")
			case accuracy >= 75:
				fmt.Println("👏 Good job! You're improving!")
			case accuracy >= 65:
				fmt.Println("📈 Nice progress! Practice makes perfect!")
			default:
				fmt.Println("💪 Keep practicing! You'll get better!")
			}
		}

		fmt.Println("\nThanks for practicing! 👋")
		os.Exit(0)
	}()
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// Set up signal handler for Ctrl+C
	setupSignalHandler()

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "help", "-h", "--help":
			printUsage()
			return
		case "stats":
			showCharacterStats()
			return
		default:
			fmt.Printf("Unknown option: %s\n", os.Args[1])
			printUsage()
			return
		}
	}

	printWelcome()
	runTypingSession()
}
