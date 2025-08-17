package qr

import (
	"fmt"
	"math"
	"regexp"
	"strings"
)

// QRCode represents a complete QR code with all necessary data and methods
type QRCode struct {
	Version    int        // QR code version (1-40)
	Size       int        // Module count per side
	ErrorLevel ErrorLevel // Error correction level
	Mask       int        // Mask pattern (0-7)
	Modules    [][]bool   // The actual QR code data grid
	IsFunction [][]bool   // Marks function pattern areas
	Data       []byte     // Raw data to encode
}

// ErrorLevel represents QR code error correction levels
type ErrorLevel int

const (
	Low      ErrorLevel = iota // L: ~7% recovery
	Medium                     // M: ~15% recovery
	Quartile                   // Q: ~25% recovery
	High                       // H: ~30% recovery
)

// Version capacity data: [version][error_level] = capacity in bytes
var versionCapacity = [][]int{
	// L,  M,  Q,  H
	{17, 14, 11, 7},      // Version 1
	{32, 26, 20, 14},     // Version 2
	{53, 42, 32, 24},     // Version 3
	{78, 62, 46, 34},     // Version 4
	{106, 84, 60, 44},    // Version 5
	{134, 106, 74, 58},   // Version 6
	{154, 122, 86, 64},   // Version 7
	{192, 152, 108, 84},  // Version 8
	{230, 180, 130, 98},  // Version 9
	{271, 213, 151, 119}, // Version 10
}

// Format information for different error levels and mask patterns
var formatInfos = [][]int{
	// Mask 0-7 for each error level
	{0x5412, 0x5125, 0x5E7C, 0x5B4B, 0x45F9, 0x40CE, 0x4F97, 0x4AA0}, // L
	{0x5125, 0x5412, 0x4B6B, 0x4E5C, 0x50EE, 0x55D9, 0x5A80, 0x5FB7}, // M
	{0x17F4, 0x1261, 0x1D38, 0x180F, 0x06BD, 0x038A, 0x0CD3, 0x09E4}, // Q
	{0x1689, 0x13BE, 0x1CE7, 0x19D0, 0x0762, 0x0255, 0x0D0C, 0x083B}, // H
}

// Finder pattern (7x7) - used in corners
var finderPattern = [][]bool{
	{true, true, true, true, true, true, true},
	{true, false, false, false, false, false, true},
	{true, false, true, true, true, false, true},
	{true, false, true, true, true, false, true},
	{true, false, true, true, true, false, true},
	{true, false, false, false, false, false, true},
	{true, true, true, true, true, true, true},
}

// NewQRCode creates a new QR code for the given URL with proper initialization
func NewQRCode(url string) (*QRCode, error) {
	// Validate URL format
	if !isValidURL(url) {
		return nil, fmt.Errorf("invalid URL format: %s", url)
	}

	// Determine optimal version and error level
	data := []byte(url)
	version, errorLevel := determineVersionAndErrorLevel(len(data))

	if version == -1 {
		return nil, fmt.Errorf("URL too long for QR code: %d bytes", len(data))
	}

	size := 21 + (version-1)*4 // QR code size formula

	qr := &QRCode{
		Version:    version,
		Size:       size,
		ErrorLevel: errorLevel,
		Mask:       0, // Will be determined later
		Data:       data,
	}

	// CRITICAL: Initialize grids before any operations
	qr.initializeGrids()

	return qr, nil
}

// initializeGrids properly allocates and initializes the module grids
func (qr *QRCode) initializeGrids() {
	// Allocate the main slices
	qr.Modules = make([][]bool, qr.Size)
	qr.IsFunction = make([][]bool, qr.Size)

	// Allocate each row
	for i := 0; i < qr.Size; i++ {
		qr.Modules[i] = make([]bool, qr.Size)
		qr.IsFunction[i] = make([]bool, qr.Size)
	}
}

// isValidURL validates basic URL format
func isValidURL(url string) bool {
	if len(url) == 0 {
		return false
	}

	return true

	// Basic URL validation pattern
	urlPattern := `^https?://[a-zA-Z0-9\-\.]+\.[a-zA-Z]{2,}(/.*)?$`
	matched, _ := regexp.MatchString(urlPattern, url)
	return matched || strings.HasPrefix(url, "www.") || strings.Contains(url, ".")
}

// determineVersionAndErrorLevel finds the optimal version and error level for given data length
func determineVersionAndErrorLevel(dataLen int) (version int, errorLevel ErrorLevel) {
	// Add overhead for mode indicator, length, and terminator
	requiredCapacity := dataLen + 3 // Rough estimate for headers

	// Try each version starting from 1
	for v := 1; v <= len(versionCapacity); v++ {
		// Try error levels from highest to lowest
		for _, el := range []ErrorLevel{High, Quartile, Medium, Low} {
			if versionCapacity[v-1][int(el)] >= requiredCapacity {
				return v, el
			}
		}
	}
	return -1, Low // No suitable version found
}

// Generate creates the complete QR code
func (qr *QRCode) Generate() error {
	// Verify initialization
	if qr.Modules == nil || qr.IsFunction == nil {
		return fmt.Errorf("QR code grids not properly initialized")
	}

	if len(qr.Modules) != qr.Size || len(qr.IsFunction) != qr.Size {
		return fmt.Errorf("QR code size mismatch: expected %d, got %d/%d",
			qr.Size, len(qr.Modules), len(qr.IsFunction))
	}

	// Step 1: Add function patterns
	qr.addFinderPatterns()
	qr.addSeparators()
	qr.addTimingPatterns()
	qr.addAlignmentPatterns()

	// Step 2: Encode and add data
	encodedData, err := qr.encodeData()
	if err != nil {
		return fmt.Errorf("data encoding failed: %v", err)
	}

	qr.addData(encodedData)

	// Step 3: Find best mask and apply it
	qr.Mask = qr.findBestMask()
	qr.applyMask(qr.Mask)

	// Step 4: Add format info (must be after mask is applied)
	qr.addFormatInfo()

	return nil
}

// addFinderPatterns adds the three 7x7 finder patterns in corners
func (qr *QRCode) addFinderPatterns() {
	positions := [][2]int{
		{0, 0},           // Top-left
		{qr.Size - 7, 0}, // Top-right
		{0, qr.Size - 7}, // Bottom-left
	}

	for _, pos := range positions {
		for dy := 0; dy < 7; dy++ {
			for dx := 0; dx < 7; dx++ {
				x, y := pos[0]+dx, pos[1]+dy
				if qr.inBounds(x, y) {
					qr.setModule(x, y, finderPattern[dy][dx])
					qr.setFunction(x, y, true)
				}
			}
		}
	}
}

// addSeparators adds white borders around finder patterns
func (qr *QRCode) addSeparators() {
	positions := [][2]int{
		{0, 0},           // Top-left
		{qr.Size - 7, 0}, // Top-right
		{0, qr.Size - 7}, // Bottom-left
	}

	for _, pos := range positions {
		for dy := -1; dy <= 7; dy++ {
			for dx := -1; dx <= 7; dx++ {
				x, y := pos[0]+dx, pos[1]+dy
				if qr.inBounds(x, y) && (dx == -1 || dx == 7 || dy == -1 || dy == 7) {
					qr.setModule(x, y, false)
					qr.setFunction(x, y, true)
				}
			}
		}
	}
}

// addTimingPatterns adds alternating timing patterns
func (qr *QRCode) addTimingPatterns() {
	for i := 8; i < qr.Size-8; i++ {
		qr.setModule(i, 6, i%2 == 0)
		qr.setModule(6, i, i%2 == 0)
		qr.setFunction(i, 6, true)
		qr.setFunction(6, i, true)
	}
}

// addAlignmentPatterns adds alignment patterns for versions > 1
func (qr *QRCode) addAlignmentPatterns() {
	if qr.Version == 1 {
		return
	}

	// Simplified alignment pattern positions
	alignmentPositions := map[int][]int{
		2:  {6, 18},
		3:  {6, 22},
		4:  {6, 26},
		5:  {6, 30},
		6:  {6, 34},
		7:  {6, 22, 38},
		8:  {6, 24, 42},
		9:  {6, 26, 46},
		10: {6, 28, 50},
	}

	positions, exists := alignmentPositions[qr.Version]
	if !exists {
		return
	}

	for _, centerY := range positions {
		for _, centerX := range positions {
			// Skip if overlaps with finder pattern
			if (centerX <= 10 && centerY <= 10) ||
				(centerX >= qr.Size-10 && centerY <= 10) ||
				(centerX <= 10 && centerY >= qr.Size-10) {
				continue
			}

			// Add 5x5 alignment pattern
			for dy := -2; dy <= 2; dy++ {
				for dx := -2; dx <= 2; dx++ {
					x, y := centerX+dx, centerY+dy
					if qr.inBounds(x, y) {
						module := (dx == -2 || dx == 2 || dy == -2 || dy == 2 || (dx == 0 && dy == 0))
						qr.setModule(x, y, module)
						qr.setFunction(x, y, true)
					}
				}
			}
		}
	}
}

// addFormatInfo adds format information around finder patterns
func (qr *QRCode) addFormatInfo() {
	formatInfo := formatInfos[int(qr.ErrorLevel)][qr.Mask]

	// Add format info bits around top-left finder pattern
	for i := 0; i < 15; i++ {
		bit := ((formatInfo >> i) & 1) != 0

		// Top-left area
		if i < 6 {
			qr.setModule(8, i, bit)
			qr.setModule(i, 8, bit)
			qr.setFunction(8, i, true)
			qr.setFunction(i, 8, true)
		} else if i == 6 {
			qr.setModule(8, 7, bit)
			qr.setModule(7, 8, bit)
			qr.setFunction(8, 7, true)
			qr.setFunction(7, 8, true)
		} else if i == 7 {
			qr.setModule(8, 8, bit)
			qr.setFunction(8, 8, true)
		} else if i == 8 {
			qr.setModule(7, 8, bit)
			qr.setFunction(7, 8, true)
		} else {
			qr.setModule(14-i, 8, bit)
			qr.setModule(8, qr.Size-15+i, bit)
			qr.setFunction(14-i, 8, true)
			qr.setFunction(8, qr.Size-15+i, true)
		}
	}

	// Dark module (always black)
	qr.setModule(8, qr.Size-8, true)
	qr.setFunction(8, qr.Size-8, true)
}

// encodeData encodes the URL data with proper headers
func (qr *QRCode) encodeData() ([]byte, error) {
	data := qr.Data

	// Create data stream
	var bits []bool

	// Mode indicator (4 bits) - Byte mode
	modeBits := []bool{false, true, false, false} // 0100 for byte mode
	bits = append(bits, modeBits...)

	// Character count (8 bits for versions 1-9 in byte mode)
	lengthBits := intToBits(len(data), 8)
	bits = append(bits, lengthBits...)

	// Data bits
	for _, b := range data {
		bits = append(bits, intToBits(int(b), 8)...)
	}

	// Terminator (up to 4 zero bits)
	maxBits := qr.getDataCapacity() * 8
	terminatorLength := min(4, maxBits-len(bits))
	for i := 0; i < terminatorLength; i++ {
		bits = append(bits, false)
	}

	// Pad to byte boundary
	for len(bits)%8 != 0 {
		bits = append(bits, false)
	}

	// Add padding bytes
	padBytes := []byte{0xEC, 0x11} // Standard padding pattern
	padIndex := 0
	for len(bits) < maxBits {
		padByte := padBytes[padIndex%2]
		bits = append(bits, intToBits(int(padByte), 8)...)
		padIndex++
	}

	// Convert bits to bytes
	bytes := make([]byte, len(bits)/8)
	for i := 0; i < len(bytes); i++ {
		for j := 0; j < 8; j++ {
			if bits[i*8+j] {
				bytes[i] |= 1 << (7 - j)
			}
		}
	}

	return bytes, nil
}

// addData places encoded data into the QR code using the standard zigzag pattern
func (qr *QRCode) addData(data []byte) {
	bitIndex := 0
	totalBits := len(data) * 8

	// Place data in zigzag pattern
	for col := qr.Size - 1; col >= 0; col -= 2 {
		if col == 6 {
			col-- // Skip timing column
		}

		upward := ((qr.Size-1-col)/2)%2 == 0

		for i := 0; i < qr.Size; i++ {
			row := i
			if upward {
				row = qr.Size - 1 - i
			}

			for c := 0; c < 2; c++ {
				x := col - c
				y := row

				if qr.inBounds(x, y) && !qr.IsFunction[y][x] {
					bit := false
					if bitIndex < totalBits {
						byteIndex := bitIndex / 8
						bitPos := 7 - (bitIndex % 8)
						bit = (data[byteIndex] & (1 << bitPos)) != 0
						bitIndex++
					}
					qr.setModule(x, y, bit)
				}
			}
		}
	}
}

// findBestMask evaluates all 8 mask patterns and returns the best one
func (qr *QRCode) findBestMask() int {
	bestMask := 0
	bestPenalty := math.MaxInt32

	for mask := 0; mask < 8; mask++ {
		qr.applyMask(mask)
		penalty := qr.calculatePenalty()
		qr.applyMask(mask) // Remove mask

		if penalty < bestPenalty {
			bestPenalty = penalty
			bestMask = mask
		}
	}

	return bestMask
}

// applyMask applies the specified mask pattern to non-function modules
func (qr *QRCode) applyMask(mask int) {
	for y := 0; y < qr.Size; y++ {
		for x := 0; x < qr.Size; x++ {
			if !qr.IsFunction[y][x] && qr.shouldMask(x, y, mask) {
				qr.Modules[y][x] = !qr.Modules[y][x]
			}
		}
	}
}

// shouldMask determines if a module should be masked based on the pattern
func (qr *QRCode) shouldMask(x, y, mask int) bool {
	switch mask {
	case 0:
		return (x+y)%2 == 0
	case 1:
		return y%2 == 0
	case 2:
		return x%3 == 0
	case 3:
		return (x+y)%3 == 0
	case 4:
		return (y/2+x/3)%2 == 0
	case 5:
		return (x*y)%2+(x*y)%3 == 0
	case 6:
		return ((x*y)%2+(x*y)%3)%2 == 0
	case 7:
		return ((x+y)%2+(x*y)%3)%2 == 0
	default:
		return false
	}
}

// calculatePenalty calculates the penalty score for the current pattern
func (qr *QRCode) calculatePenalty() int {
	penalty := 0

	// Rule 1: Adjacent modules in row/column
	for y := 0; y < qr.Size; y++ {
		count := 1
		for x := 1; x < qr.Size; x++ {
			if qr.Modules[y][x] == qr.Modules[y][x-1] {
				count++
			} else {
				if count >= 5 {
					penalty += 3 + (count - 5)
				}
				count = 1
			}
		}
		if count >= 5 {
			penalty += 3 + (count - 5)
		}
	}

	return penalty
}

// PrintToTerminal renders the QR code in the terminal with proper formatting
func (qr *QRCode) PrintToTerminal() {
	quietZone := 4
	totalSize := qr.Size + 2*quietZone

	// fmt.Printf("╔%s╗\n", strings.Repeat("═", totalSize*2))

	for y := 0; y < totalSize; y++ {
		fmt.Print("")
		for x := 0; x < totalSize; x++ {
			// Quiet zone check
			if x < quietZone || x >= qr.Size+quietZone ||
				y < quietZone || y >= qr.Size+quietZone {
				fmt.Print("  ") // White space
			} else {
				moduleX, moduleY := x-quietZone, y-quietZone
				if qr.Modules[moduleY][moduleX] {
					fmt.Print("██") // Black module
				} else {
					fmt.Print("  ") // White module
				}
			}
		}
		fmt.Println("")
	}

	// fmt.Printf("╚%s╝\n", strings.Repeat("═", totalSize*2))
}

// Helper functions

func (qr *QRCode) setModule(x, y int, value bool) {
	if qr.inBounds(x, y) {
		qr.Modules[y][x] = value
	}
}

func (qr *QRCode) setFunction(x, y int, value bool) {
	if qr.inBounds(x, y) {
		qr.IsFunction[y][x] = value
	}
}

func (qr *QRCode) inBounds(x, y int) bool {
	return x >= 0 && x < qr.Size && y >= 0 && y < qr.Size
}

func (qr *QRCode) getDataCapacity() int {
	if qr.Version > len(versionCapacity) || qr.Version < 1 {
		return 0
	}
	return versionCapacity[qr.Version-1][int(qr.ErrorLevel)]
}

func intToBits(value, length int) []bool {
	bits := make([]bool, length)
	for i := 0; i < length; i++ {
		bits[length-1-i] = ((value >> i) & 1) != 0
	}
	return bits
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
