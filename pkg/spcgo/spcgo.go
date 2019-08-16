package spcgo

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"os"
)

// Header stores header info for SPC file
type Header struct {
	Ftflg, Fversn, Fexper, Fexp   uint8
	Fnpts                         int32
	Ffirst, Flast                 float64
	Fnsub                         int32
	Fxtype, Fytype, Fztype, Fpost uint8
	Fdate                         int32
	Fres, Fsource                 [9]byte
	Fpeakpt                       int16
	Fspare                        [32]uint8
	Fcmnt                         [130]uint8
	Fcatxt                        [30]uint8
	Flogoff                       int32
	Fmods                         int32
	Fprocs, Flevel                uint8
	Fsampin                       int16
	Ffactor                       float32
	Fmethod                       [48]uint8
	Fzinc                         float32
	Fwplanes                      int32
	Fwinc                         float32
	Fwtype                        uint8
	Freserv                       [187]uint8
}

// Date is a more usable date format for SPC file
type Date struct {
	Year, Month, Day, Hour, Minute int32
}

// SubHeader stores sub header for SPC file
type SubHeader struct {
	Sflags, Sexp           uint8
	Sindex                 uint16
	Sstartz, Sendz, Snoise float32
	Snpts, Snscans         int32
	Swaxis                 float32
	Sreserv                [4]uint8
}

// Flags store flags for SPC file
type Flags struct {
	Tsprec, Tcgram, Tmulti, Trandm, Tordrd, Talabs, Txyxys, Txvals bool
}

// SubFlags stores sub flags for SPC file
type SubFlags struct {
	Tchng, Tdnupk, Tmodar bool
}

// Data stores data for SPC file
type Data struct {
	X, Y *[]float32
}

func (d Data) length() uint64 {
	if len(*d.X) == len(*d.Y) {
		return uint64(len(*d.X))
	}
	return 0
}

func (d Data) stringCSV(index uint64) string {
	//if index < d.length() {
	return fmt.Sprintf("%f,%f", (*d.X)[index], (*d.Y)[index])
	//}
	//return ""
}

// LogHeader stores the log header for SPC file
type LogHeader struct {
	Lsize, Lspace, Loff, Lbnsz, Lbnspc uint32
	Lreserv                            [44]byte
}

// SPCfile stores required parts for SPC file
type SPCfile struct {
	Head  *Header
	SHead *SubHeader
	LHead *LogHeader
	Data  *Data
}

// HeaderReader read header out of spc file
func HeaderReader(content []byte) *Header {
	R := bytes.NewReader(content[0:512])
	var head Header
	if err := binary.Read(R, binary.LittleEndian, &head); err != nil {
		fmt.Println("binary.Read failed:", err)
		os.Exit(3)
	}
	return &head
}

// SubHeaderReader reads sub header out of spc file
func SubHeaderReader(content []byte, start *int32) *SubHeader {
	R := bytes.NewReader(content[*start : *start+32])
	*start += 32
	var sHead SubHeader
	if err := binary.Read(R, binary.LittleEndian, &sHead); err != nil {
		fmt.Println("binary.Read failed:", err)
		os.Exit(3)
	}
	return &sHead
}

// LogHeaderReader reads the log header out of spc file
func LogHeaderReader(content []byte, start *int32) *LogHeader {
	R := bytes.NewReader(content[*start : *start+64])
	*start += 64
	var lHead LogHeader
	if err := binary.Read(R, binary.LittleEndian, &lHead); err != nil {
		fmt.Println("binary.Read failed:", err)
		os.Exit(3)
	}
	return &lHead
}

// FlagsUnpack unpacks flags from flag variable
func FlagsUnpack(Ftflg uint8, verbose bool) *Flags {
	var Fflags Flags
	Fflags.Tsprec = (Ftflg>>0)&1 == 1
	Fflags.Tcgram = (Ftflg>>1)&1 == 1
	Fflags.Tmulti = (Ftflg>>2)&1 == 1
	Fflags.Trandm = (Ftflg>>3)&1 == 1
	Fflags.Tordrd = (Ftflg>>4)&1 == 1
	Fflags.Talabs = (Ftflg>>5)&1 == 1
	Fflags.Txyxys = (Ftflg>>6)&1 == 1
	Fflags.Txvals = (Ftflg>>7)&1 == 1
	// Report on flags
	if verbose {
		fmt.Printf("Flags variable: %d\n", Ftflg)
		if Fflags.Tsprec {
			fmt.Printf("Y data is stored in 16-bit precision (instead of 32-bit)\n")
		}
		if Fflags.Tcgram {
			fmt.Printf("Use Experiment extension, not SPC\n")
		}
		if Fflags.Tmulti {
			fmt.Printf("Multifile\n")
		}
		if Fflags.Trandm {
			fmt.Printf("If a Multifile, Z values are randomly ordered\n")
		}
		if Fflags.Tordrd {
			fmt.Printf("If a Multifile, Z values are ordered, but not even\n")
		}
		if Fflags.Talabs {
			fmt.Printf("Use custom axis labels (obsolete)\n")
		}
		if Fflags.Txyxys {
			fmt.Printf("If an XY file and a Multifile, each subfile has its own X array\n")
		} else if Fflags.Txvals {
			fmt.Printf("XY file\n")
		} else {
			fmt.Printf("No x given, must be generated\n")
		}
	}
	return &Fflags
}

// SubFlagsUnpack Unpacks sub flags from sub flag variable
func SubFlagsUnpack(Ftflg uint8, verbose bool) *SubFlags {
	var Sflags SubFlags
	Sflags.Tchng = (Ftflg>>0)&1 == 1 // need to verify bit locations
	Sflags.Tdnupk = (Ftflg>>3)&1 == 1
	Sflags.Tmodar = (Ftflg>>7)&1 == 1
	// Report on flags
	if verbose {
		if Sflags.Tchng {
			fmt.Printf("Subfiles changed\n")
		}
		if Sflags.Tdnupk {
			fmt.Printf("Do not use peak table file\n")
		}
		if Sflags.Tmodar {
			fmt.Printf("Subfile modified by arithmetic\n")
		}
	}
	return &Sflags
}

// DateUnpack Unpacks date from int32
func DateUnpack(Fdate int32, verbose bool) *Date {
	var Date Date
	Date.Year = Fdate >> 20
	Date.Month = (Fdate >> 16) & 31 //(2 << 4 - 1)) //% (2 << 4))
	Date.Day = (Fdate >> 11) & 63   //(2 << 5 - 1)) //% (2 << 5))
	Date.Hour = (Fdate >> 6) & 63   //(2 << 5 - 1)) //% (2 << 5))
	Date.Minute = Fdate & 63        //(2 << 5 - 1)) //% (2 << 6))
	if verbose {
		fmt.Printf("Year: %d\nMonth: %d\nDay: %d\nHour: %d\nMinute: %d\n", Date.Year, Date.Month, Date.Day, Date.Hour, Date.Minute)
	}
	return &Date
}

// DatePack Packs date into int32
func DatePack(Date *Date, verbose bool) int32 {
	var output int32
	output = Date.Year << 20
	output = output + Date.Month<<16
	output = output + Date.Day<<11
	output = output + Date.Hour<<6
	output = output + Date.Minute
	if verbose {
		fmt.Printf("Date packed into %d.\n", output)
	}
	return output
}

// SPCPack packs an SPC object into a binary File struct
func SPCPack(head *Header, sHead *SubHeader, lHead *LogHeader, data *Data, verbose bool) *bytes.Buffer {
	head.Fexp = uint8(128)
	head.Ftflg = uint8(128)
	head.Flogoff = int32(0)
	/*
		lHead.Lsize = uint32(0)
		lHead.Lspace = uint32(0)
		lHead.Loff = uint32(0)
		lHead.Lbnsz = uint32(0)
		lHead.Lbnspc = uint32(0)
	*/
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, head)
	binary.Write(&buf, binary.LittleEndian, data.X)
	binary.Write(&buf, binary.LittleEndian, sHead)
	binary.Write(&buf, binary.LittleEndian, data.Y)
	//binary.Write(&buf, binary.LittleEndian, lHead)
	return &buf
}

func saveBuffer(buf *bytes.Buffer, filename string, verbose bool) error {
	file, err := os.Create(filename)
	defer file.Close()

	if err != nil {
		return err
	}
	_, err = file.Write(buf.Bytes())

	if err != nil {
		return err
	}
	return nil
}

// SaveSPC saves the SPC file
func SaveSPC(SPC SPCfile, filename string, verbose bool) {
	buf := SPCPack(SPC.Head, SPC.SHead, SPC.LHead, SPC.Data, verbose)
	err := saveBuffer(buf, filename, verbose)

	if err != nil {
		log.Fatal(err)
		os.Exit(4)
	}
}

// SaveCSV saves a CSV file
func SaveCSV(SPC SPCfile, filename string) {
	csvFile, err := os.Create(filename)
	if err != nil {
		log.Fatalf("could not open csv file to write | %v\n", err)
	}
	defer csvFile.Close()

	for i := uint64(0); i < SPC.Data.length(); i++ {
		fmt.Fprintln(csvFile, SPC.Data.stringCSV(i))
	}
}

// linespace(start, stop, num=50, endpoint=True, retstep=False, dtype=None)[source] Code taken from pa-m/numgo.
func linespace(start, stop float32, num int32, endPoint bool) []float32 {
	step := float32(0)
	if endPoint {
		if num == 1 {
			return []float32{start}
		}
		step = (stop - start) / float32(num-1)
	} else {
		if num == 0 {
			return []float32{}
		}
		step = (stop - start) / float32(num)
	}
	r := make([]float32, num, num)
	for i := 0; i < int(num); i++ {
		r[i] = start + float32(i)*step
	}
	return r
}

func readBin(filename string) ([]byte, int64, error) {
	file, err := os.Open(filename)

	if err != nil {
		return nil, 0, err
	}
	defer file.Close()

	stats, statsErr := file.Stat()
	if statsErr != nil {
		return nil, 0, statsErr
	}

	size := stats.Size()
	content := make([]byte, size)

	bufr := bufio.NewReader(file)
	_, err = bufr.Read(content)

	return content, size, err
}

// ReadSPC takes in a filename and verbose boolean and returns an SPCfile struct
func ReadSPC(filename string, verbose bool) SPCfile {
	// open file
	content, size, _ := readBin(filename)

	head := HeaderReader(content)
	var SPC SPCfile
	SPC.Head = head
	/* Test date unpacking/packing
	upackedDate := dateUnpack(head.Fdate, verbose)
	repackedDate := datePack(upackedDate, verbose)
	dateUnpack(repackedDate, verbose)
	*/
	var filePos int32 = 512
	if verbose {
		fmt.Printf("File is %d bytes long.\n", size)
		fmt.Printf("File contains %d points.\n", head.Fnpts)
		fmt.Printf("File exponent is %d.\n", head.Fexp)
		fmt.Printf("File contains %d spectra.\n", head.Fnsub)
		fmt.Printf("Y type is %d.\n", head.Fytype)
		fmt.Printf("Log offset is %d.\n", head.Flogoff)
	}
	Fflags := FlagsUnpack(head.Ftflg, verbose)

	if !Fflags.Tmulti {
		x := make([]float32, head.Fnpts)
		y := make([]float32, head.Fnpts)
		spcData := Data{X: &x, Y: &y}
		SPC.Data = &spcData
		if Fflags.Txvals {
			if verbose {
				fmt.Printf("Single spectra file with included X.\n")
			}
			fmt.Printf("start X: %f\nend x: %f\n", head.Ffirst, head.Flast)

			r := bytes.NewReader(content[filePos:(filePos + head.Fnpts*4)])
			filePos = filePos + head.Fnpts*4
			if err := binary.Read(r, binary.LittleEndian, spcData.X); err != nil {
				fmt.Println("binary.Read failed:", err)
				os.Exit(3)
			}
		} else {
			if verbose {
				fmt.Printf("Single spectra file with generated X.\n")
			}
			x = linespace(float32(head.Ffirst), float32(head.Flast), head.Fnpts, true)
		}
		subHead := SubHeaderReader(content, &filePos)
		Sflags := SubFlagsUnpack(subHead.Sflags, verbose)
		_ = Sflags
		SPC.SHead = subHead

		if verbose {
			fmt.Printf("File position updated to %d.\n", filePos)
		}
		r := bytes.NewReader(content[filePos:(filePos + head.Fnpts*4)])
		filePos = filePos + head.Fnpts*4
		if verbose {
			fmt.Printf("File position updated to %d.\n", filePos)
		}
		if head.Fexp == 128 {
			if err := binary.Read(r, binary.LittleEndian, spcData.Y); err != nil {
				fmt.Println("binary.Read failed:", err)
				os.Exit(3)
			}
		} else {
			dataRead := make([]int32, head.Fnpts)
			var factor float32
			if Fflags.Tsprec {
				factor = float32(math.Pow(2, float64(head.Fexp)-16))
			} else {
				factor = float32(math.Pow(2, float64(head.Fexp)-32))
			}
			if err := binary.Read(r, binary.LittleEndian, &dataRead); err != nil {
				fmt.Println("binary.Read failed:", err)
				os.Exit(3)
			}
			for i := range dataRead {
				y[i] = float32(dataRead[i]) * factor
			}
		}
	} else {
		fmt.Printf("Multiple spectra files not implemented yet.\n")
	}
	if head.Flogoff != 0 {
		LHead := LogHeaderReader(content, &filePos)
		SPC.LHead = LHead
		fmt.Printf("Lsize: %d\nLspace: %d\nLoff: %d\nLbnsz: %d\nLbnspc: %d\n", LHead.Lsize, LHead.Lspace, LHead.Loff, LHead.Lbnsz, LHead.Lbnspc)
	}
	return SPC
}
