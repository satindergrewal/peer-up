// Derived from github.com/skip2/go-qrcode (MIT License).
// Copyright (c) 2014 Tom Harwood. See THIRD_PARTY_NOTICES in the repo root.
//
// Original: https://github.com/skip2/go-qrcode
// Modifications: merged symbol.go and regular_symbol.go; uses package-local
// bitset type; removed unused string() method.

package qr

// Abbreviated true/false for pattern tables.
const (
	b0 = false
	b1 = true
)

// symbol is a 2D array of bits representing a QR Code symbol.
type symbol struct {
	module     [][]bool
	isUsed     [][]bool
	size       int
	symbolSize int
	quietZone  int
}

func newSymbol(size int, quietZone int) *symbol {
	m := &symbol{
		module:     make([][]bool, size+2*quietZone),
		isUsed:     make([][]bool, size+2*quietZone),
		size:       size + 2*quietZone,
		symbolSize: size,
		quietZone:  quietZone,
	}
	for i := range m.module {
		m.module[i] = make([]bool, size+2*quietZone)
		m.isUsed[i] = make([]bool, size+2*quietZone)
	}
	return m
}

func (m *symbol) get(x int, y int) bool {
	return m.module[y+m.quietZone][x+m.quietZone]
}

func (m *symbol) empty(x int, y int) bool {
	return !m.isUsed[y+m.quietZone][x+m.quietZone]
}

func (m *symbol) numEmptyModules() int {
	count := 0
	for y := 0; y < m.symbolSize; y++ {
		for x := 0; x < m.symbolSize; x++ {
			if !m.isUsed[y+m.quietZone][x+m.quietZone] {
				count++
			}
		}
	}
	return count
}

func (m *symbol) set(x int, y int, v bool) {
	m.module[y+m.quietZone][x+m.quietZone] = v
	m.isUsed[y+m.quietZone][x+m.quietZone] = true
}

func (m *symbol) set2dPattern(x int, y int, v [][]bool) {
	for j, row := range v {
		for i, value := range row {
			m.set(x+i, y+j, value)
		}
	}
}

func (m *symbol) bitmap() [][]bool {
	module := make([][]bool, len(m.module))
	for i := range m.module {
		module[i] = m.module[i][:]
	}
	return module
}

// Penalty scoring constants (ISO/IEC 18004:2006).
const (
	penaltyWeight1 = 3
	penaltyWeight2 = 3
	penaltyWeight3 = 40
	penaltyWeight4 = 10
)

func (m *symbol) penaltyScore() int {
	return m.penalty1() + m.penalty2() + m.penalty3() + m.penalty4()
}

func (m *symbol) penalty1() int {
	penalty := 0
	for x := 0; x < m.symbolSize; x++ {
		lastValue := m.get(x, 0)
		count := 1
		for y := 1; y < m.symbolSize; y++ {
			v := m.get(x, y)
			if v != lastValue {
				count = 1
				lastValue = v
			} else {
				count++
				if count == 6 {
					penalty += penaltyWeight1 + 1
				} else if count > 6 {
					penalty++
				}
			}
		}
	}
	for y := 0; y < m.symbolSize; y++ {
		lastValue := m.get(0, y)
		count := 1
		for x := 1; x < m.symbolSize; x++ {
			v := m.get(x, y)
			if v != lastValue {
				count = 1
				lastValue = v
			} else {
				count++
				if count == 6 {
					penalty += penaltyWeight1 + 1
				} else if count > 6 {
					penalty++
				}
			}
		}
	}
	return penalty
}

func (m *symbol) penalty2() int {
	penalty := 0
	for y := 1; y < m.symbolSize; y++ {
		for x := 1; x < m.symbolSize; x++ {
			topLeft := m.get(x-1, y-1)
			above := m.get(x, y-1)
			left := m.get(x-1, y)
			current := m.get(x, y)
			if current == left && current == above && current == topLeft {
				penalty++
			}
		}
	}
	return penalty * penaltyWeight2
}

func (m *symbol) penalty3() int {
	penalty := 0
	for y := 0; y < m.symbolSize; y++ {
		var bitBuffer int16
		for x := 0; x < m.symbolSize; x++ {
			bitBuffer <<= 1
			if v := m.get(x, y); v {
				bitBuffer |= 1
			}
			switch bitBuffer & 0x7ff {
			case 0x05d, 0x5d0:
				penalty += penaltyWeight3
				bitBuffer = 0xFF
			default:
				if x == m.symbolSize-1 && (bitBuffer&0x7f) == 0x5d {
					penalty += penaltyWeight3
					bitBuffer = 0xFF
				}
			}
		}
	}
	for x := 0; x < m.symbolSize; x++ {
		var bitBuffer int16
		for y := 0; y < m.symbolSize; y++ {
			bitBuffer <<= 1
			if v := m.get(x, y); v {
				bitBuffer |= 1
			}
			switch bitBuffer & 0x7ff {
			case 0x05d, 0x5d0:
				penalty += penaltyWeight3
				bitBuffer = 0xFF
			default:
				if y == m.symbolSize-1 && (bitBuffer&0x7f) == 0x5d {
					penalty += penaltyWeight3
					bitBuffer = 0xFF
				}
			}
		}
	}
	return penalty
}

func (m *symbol) penalty4() int {
	numModules := m.symbolSize * m.symbolSize
	numDarkModules := 0
	for x := 0; x < m.symbolSize; x++ {
		for y := 0; y < m.symbolSize; y++ {
			if v := m.get(x, y); v {
				numDarkModules++
			}
		}
	}
	numDarkModuleDeviation := numModules/2 - numDarkModules
	if numDarkModuleDeviation < 0 {
		numDarkModuleDeviation *= -1
	}
	return penaltyWeight4 * (numDarkModuleDeviation / (numModules / 20))
}

// --- Regular symbol construction ---

var (
	alignmentPatternCenter = [][]int{
		{}, {}, {6, 18}, {6, 22}, {6, 26}, {6, 30}, {6, 34},
		{6, 22, 38}, {6, 24, 42}, {6, 26, 46}, {6, 28, 50},
		{6, 30, 54}, {6, 32, 58}, {6, 34, 62}, {6, 26, 46, 66},
		{6, 26, 48, 70}, {6, 26, 50, 74}, {6, 30, 54, 78},
		{6, 30, 56, 82}, {6, 30, 58, 86}, {6, 34, 62, 90},
		{6, 28, 50, 72, 94}, {6, 26, 50, 74, 98}, {6, 30, 54, 78, 102},
		{6, 28, 54, 80, 106}, {6, 32, 58, 84, 110}, {6, 30, 58, 86, 114},
		{6, 34, 62, 90, 118}, {6, 26, 50, 74, 98, 122},
		{6, 30, 54, 78, 102, 126}, {6, 26, 52, 78, 104, 130},
		{6, 30, 56, 82, 108, 134}, {6, 34, 60, 86, 112, 138},
		{6, 30, 58, 86, 114, 142}, {6, 34, 62, 90, 118, 146},
		{6, 30, 54, 78, 102, 126, 150}, {6, 24, 50, 76, 102, 128, 154},
		{6, 28, 54, 80, 106, 132, 158}, {6, 32, 58, 84, 110, 136, 162},
		{6, 26, 54, 82, 110, 138, 166}, {6, 30, 58, 86, 114, 142, 170},
	}

	finderPattern = [][]bool{
		{b1, b1, b1, b1, b1, b1, b1},
		{b1, b0, b0, b0, b0, b0, b1},
		{b1, b0, b1, b1, b1, b0, b1},
		{b1, b0, b1, b1, b1, b0, b1},
		{b1, b0, b1, b1, b1, b0, b1},
		{b1, b0, b0, b0, b0, b0, b1},
		{b1, b1, b1, b1, b1, b1, b1},
	}

	finderPatternSize = 7

	finderPatternHorizontalBorder = [][]bool{{b0, b0, b0, b0, b0, b0, b0, b0}}
	finderPatternVerticalBorder   = [][]bool{{b0}, {b0}, {b0}, {b0}, {b0}, {b0}, {b0}, {b0}}

	alignmentPattern = [][]bool{
		{b1, b1, b1, b1, b1},
		{b1, b0, b0, b0, b1},
		{b1, b0, b1, b0, b1},
		{b1, b0, b0, b0, b1},
		{b1, b1, b1, b1, b1},
	}
)

type regularSymbol struct {
	version qrCodeVersion
	mask    int
	data    *bitset
	symbol  *symbol
	size    int
}

type direction uint8

const (
	up direction = iota
	down
)

func buildRegularSymbol(version qrCodeVersion, mask int, data *bitset, includeQuietZone bool) (*symbol, error) {
	quietZone := 0
	if includeQuietZone {
		quietZone = version.quietZoneSize()
	}
	m := &regularSymbol{
		version: version,
		mask:    mask,
		data:    data,
		symbol:  newSymbol(version.symbolSize(), quietZone),
		size:    version.symbolSize(),
	}
	m.addFinderPatterns()
	m.addAlignmentPatterns()
	m.addTimingPatterns()
	m.addFormatInfo()
	m.addVersionInfo()
	ok, err := m.addData()
	if !ok {
		return nil, err
	}
	return m.symbol, nil
}

func (m *regularSymbol) addFinderPatterns() {
	fpSize := finderPatternSize
	m.symbol.set2dPattern(0, 0, finderPattern)
	m.symbol.set2dPattern(0, fpSize, finderPatternHorizontalBorder)
	m.symbol.set2dPattern(fpSize, 0, finderPatternVerticalBorder)
	m.symbol.set2dPattern(m.size-fpSize, 0, finderPattern)
	m.symbol.set2dPattern(m.size-fpSize-1, fpSize, finderPatternHorizontalBorder)
	m.symbol.set2dPattern(m.size-fpSize-1, 0, finderPatternVerticalBorder)
	m.symbol.set2dPattern(0, m.size-fpSize, finderPattern)
	m.symbol.set2dPattern(0, m.size-fpSize-1, finderPatternHorizontalBorder)
	m.symbol.set2dPattern(fpSize, m.size-fpSize-1, finderPatternVerticalBorder)
}

func (m *regularSymbol) addAlignmentPatterns() {
	for _, x := range alignmentPatternCenter[m.version.version] {
		for _, y := range alignmentPatternCenter[m.version.version] {
			if !m.symbol.empty(x, y) {
				continue
			}
			m.symbol.set2dPattern(x-2, y-2, alignmentPattern)
		}
	}
}

func (m *regularSymbol) addTimingPatterns() {
	value := true
	for i := finderPatternSize + 1; i < m.size-finderPatternSize; i++ {
		m.symbol.set(i, finderPatternSize-1, value)
		m.symbol.set(finderPatternSize-1, i, value)
		value = !value
	}
}

func (m *regularSymbol) addFormatInfo() {
	fpSize := finderPatternSize
	l := formatInfoLengthBits - 1
	f := m.version.formatInfo(m.mask)
	for i := 0; i <= 7; i++ {
		m.symbol.set(m.size-i-1, fpSize+1, f.at(l-i))
	}
	for i := 0; i <= 5; i++ {
		m.symbol.set(fpSize+1, i, f.at(l-i))
	}
	m.symbol.set(fpSize+1, fpSize, f.at(l-6))
	m.symbol.set(fpSize+1, fpSize+1, f.at(l-7))
	m.symbol.set(fpSize, fpSize+1, f.at(l-8))
	for i := 9; i <= 14; i++ {
		m.symbol.set(14-i, fpSize+1, f.at(l-i))
	}
	for i := 8; i <= 14; i++ {
		m.symbol.set(fpSize+1, m.size-fpSize+i-8, f.at(l-i))
	}
	m.symbol.set(fpSize+1, m.size-fpSize-1, true)
}

func (m *regularSymbol) addVersionInfo() {
	fpSize := finderPatternSize
	v := m.version.versionInfo()
	if v == nil {
		return
	}
	l := versionInfoLengthBits - 1
	for i := 0; i < v.len(); i++ {
		m.symbol.set(i/3, m.size-fpSize-4+i%3, v.at(l-i))
		m.symbol.set(m.size-fpSize-4+i%3, i/3, v.at(l-i))
	}
}

func (m *regularSymbol) addData() (bool, error) {
	xOffset := 1
	dir := up
	x := m.size - 2
	y := m.size - 1

	for i := 0; i < m.data.len(); i++ {
		var mask bool
		switch m.mask {
		case 0:
			mask = (y+x+xOffset)%2 == 0
		case 1:
			mask = y%2 == 0
		case 2:
			mask = (x+xOffset)%3 == 0
		case 3:
			mask = (y+x+xOffset)%3 == 0
		case 4:
			mask = (y/2+(x+xOffset)/3)%2 == 0
		case 5:
			mask = (y*(x+xOffset))%2+(y*(x+xOffset))%3 == 0
		case 6:
			mask = ((y*(x+xOffset))%2+((y*(x+xOffset))%3))%2 == 0
		case 7:
			mask = ((y+x+xOffset)%2+((y*(x+xOffset))%3))%2 == 0
		}
		m.symbol.set(x+xOffset, y, mask != m.data.at(i))

		if i == m.data.len()-1 {
			break
		}

		for {
			if xOffset == 1 {
				xOffset = 0
			} else {
				xOffset = 1
				if dir == up {
					if y > 0 {
						y--
					} else {
						dir = down
						x -= 2
					}
				} else {
					if y < m.size-1 {
						y++
					} else {
						dir = up
						x -= 2
					}
				}
			}
			if x == 5 {
				x--
			}
			if m.symbol.empty(x+xOffset, y) {
				break
			}
		}
	}
	return true, nil
}
