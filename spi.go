package main

import (
	"fmt"
	"time"

	"github.com/davecheney/gpio"
)

// AdcRead represents the data needed to perform a read operation on the ADC component
type AdcRead struct {
	Cs          gpio.Pin
	Clock       gpio.Pin
	Miso        gpio.Pin
	NumBits     int
	ResultsChan chan uint32
}

// Exec reads the current value stored in the ADC register
func (reader AdcRead) Exec() {

	// Start the CS Low to begin the read
	reader.Cs.Clear()

	// Initialize an impty uint32 to store the value we are reading
	var result uint32

	// Loop over each bit the the component sends back (The number depends varies from
	// component to component, read the datasheet)
	for i := 0; i < reader.NumBits; i++ {

		// Set the clock to logic high
		reader.Clock.Set()

		// Read 1 bit in, if it is high, then add a "1" to our rightmost bit
		bit := reader.Miso.Get()
		if bit {
			result |= 0x1
		}

		// Shift Left to get to the next bit to be read
		if i != reader.NumBits-1 {
			result <<= 1
		}

		// The clock will pulse low, then high again to get the next bit
		reader.Clock.Clear()
	}

	// Set chip select low to end the read
	reader.Cs.Set()

	// Send the result back through the channel to whatever part of our
	// application cares about it
	reader.ResultsChan <- result
}

func main() {

	// Open necessary pins. The numbers here are examples, they should be changed based
	// on which pins you use
	const csPinNumber = 5
	const clockPinNumber = 5
	const misoPinNumber = 5

	csPin, err := gpio.OpenPin(csPinNumber, gpio.ModeOutput)
	if err != nil {
		fmt.Printf("Error opening cs pin: %v\n", err)
	}
	clockPin, err := gpio.OpenPin(clockPinNumber, gpio.ModeOutput)
	if err != nil {
		fmt.Printf("Error opening clock pin: %v\n", err)
	}
	misoPin, err := gpio.OpenPin(misoPinNumber, gpio.ModeInput)
	if err != nil {
		fmt.Printf("Error opening miso pin: %v\n", err)
	}

	resultsChan := make(chan uint32, 1)

	adcReader := AdcRead{
		Cs:          csPin,
		Clock:       clockPin,
		Miso:        misoPin,
		NumBits:     32, // Our ADC component sends a 32 bit value
		ResultsChan: resultsChan,
	}

	// Execute each read at 10 Hz
	c := time.Tick(time.Duration(100) * time.Millisecond)
	go func() {
		for range c {
			adcReader.Exec()
		}
	}()

	// Print everything that comes through the results channel
	for true {
		result := <-resultsChan
		fmt.Println(result)
	}
}
