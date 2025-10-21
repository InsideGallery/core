package simdict

import (
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/InsideGallery/core/utils"
)

func TestManager(t *testing.T) {
	// A stream of incoming emails
	emailStream := []string{
		"test@gmail.com",        // Starts a new bucket
		"john.doe@example.com",  // Starts a new bucket
		"tset@gmail.com",        // Should be bucketed with test@gmail.com
		"jane.smith@work.net",   // Starts a new bucket
		"maxim@weavers.team",    // Starts a new bucket
		"test@gamil.com",        // Should be bucketed with test@gmail.com
		"john_doe@example.com",  // Should be bucketed with john.doe...
		"user-1234@hotmail.com", // Starts a new bucket
		"test@gmail.com",        // A duplicate, should be handled gracefully
		"maskym@weavers.team",   // Should be bucketed with maxim@weavers.team
		"maksym@weavers.team",   // Should be bucketed with maxim@weavers.team
		"maksym@gmail.com",      // Should be bucketed with maksym@gmail.com
	}

	manager := NewLSHManager()

	fmt.Println("--- Processing Email Stream ---")

	for _, email := range emailStream {
		res, err := Normalize(email)
		if err != nil {
			slog.Error("error normalize email", "email", email, "err", err)
			return
		}

		bucketID := manager.ProcessAndAssign(res)
		fmt.Printf("Email: '%-25s' -> Assigned to Bucket: '%s'\n", email, bucketID)
	}

	fmt.Println("\n--- Final Buckets ---")

	for bucketID, members := range manager.bucketMembers {
		fmt.Printf("Bucket '%s':\n", bucketID)

		for _, member := range members {
			fmt.Printf("  - %s\n", member)
		}
	}

	// A stream of incoming emails
	nameStream := [][2]string{
		{"Maksym", "Tkach"}, // Starts a new bucket
		{"Maskym", "Tkach"}, // Starts a new bucket
		{"Test", "Tkach"},   // Starts a new bucket
		{"Test", "Rose"},    // Starts a new bucket
		{"Tist", "Rose"},    // Starts a new bucket
		{"Bob", "Dilan"},    // Starts a new bucket
		{"Bob", "Delan"},    // Starts a new bucket
		{"Bab", "Delan"},    // Starts a new bucket
		{"John", "Smith"},   // Starts a new bucket
		{"John", "Smith"},
		{"Catherine", "Jones"},
		{"François", "Hollande"},
		{"John F.", "Smith"},
		{"Francoise", "Hollande"},
		{"Robert", "Smith"},
		{"Bob", "Smith"},
		{"Robert", "Smiht"}, // Starts a new bucket
		{"Maskym", "Tkoch"},
		{"Maskym", "Tkahc"},
	}

	manager = NewLSHManager()

	fmt.Println("--- Processing Names Stream ---")

	for _, name := range nameStream {
		res, err := Normalize(utils.CommonString(strings.Join(name[:], " ")))
		if err != nil {
			slog.Error("error normalize name", "name", name, "err", err)
			return
		}

		bucketID := manager.ProcessAndAssign(res)
		fmt.Printf("Name: '%-7s %-7s' -> Assigned to Bucket: '%s'\n", name[0], name[1], bucketID)
	}

	fmt.Println("\n--- Final Buckets ---")

	for bucketID, members := range manager.bucketMembers {
		fmt.Printf("Bucket '%s':\n", bucketID)

		for _, member := range members {
			fmt.Printf("  - %s\n", member)
		}
	}
}
