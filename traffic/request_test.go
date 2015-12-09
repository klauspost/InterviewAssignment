package traffic

import (
	"math/rand"
	"testing"
	"time"

	"github.com/google/gofuzz"
)

// Return a time no later than 5000 years from unix datum.
// JSON cannot handle dates after year 9999.
func fuzzTime(t *time.Time, c fuzz.Continue) {
	sec := c.Rand.Int63()
	nsec := c.Rand.Int63()
	// No more than 5000 years in the future
	sec %= 5000 * 365 * 24 * 60 * 60
	*t = time.Unix(sec, nsec)
}

// Test if hash generation is deterministic
// We do 10 runs with different data.
func TestGenerateHash(t *testing.T) {
	for i := int64(0); i < 10; i++ {
		var a, b Request

		// Fill requests with values
		f := fuzz.New()
		f.NumElements(0, 50) // Slioes have between 0 and 50 elements
		f.NilChance(0.1)     // Nilable types have 10% chance of being nil
		// Be sure we don't generate invalid times.
		f.Funcs(fuzzTime)
		f.RandSource(rand.New(rand.NewSource(i)))
		f.Fuzz(&a)
		a.GenerateHash()

		// Reset random seed
		f.RandSource(rand.New(rand.NewSource(i)))
		f.Fuzz(&b)
		b.GenerateHash()
		if a.ID == "" {
			t.Fatal("hash was not set")
		}
		if a.ID != b.ID {
			t.Fatalf("Hash was not deterministic, %q != %q", a.ID, b.ID)
		}
		if len(a.ID) != 20*2 {
			t.Fatalf("unexpected hash length, was %d, expected 40", len(a.ID))
		}
	}
}
