package main

import (
	"bytes"
	"debug/pe"
	"fmt"
	"hash/crc32"
	"io/ioutil"
	"os"

	"golang.org/x/arch/x86/x86asm"
)

const steamPath = "C:\\Program Files (x86)\\Steam\\SteamUI.dll"

const validChecksum = 0x85ac72fb

// Section for for DLL info
type Section struct {
	section *pe.Section
	arch    int
	base    uint64
	rdata   *RdataRange
}

// RdataRange provides the offset range for the rdata section
type RdataRange struct {
	StartOffset, EndOffset uint64
}

func getTextSection(buf []byte) (*Section, error) {
	reader := bytes.NewReader(buf)

	file, err := pe.NewFile(reader)
	if err != nil {
		return nil, err
	}

	var arch int
	switch file.FileHeader.Machine {
	case pe.IMAGE_FILE_MACHINE_I386:
		arch = 32
	case pe.IMAGE_FILE_MACHINE_AMD64:
		arch = 64
	default:
		return nil, fmt.Errorf("support for machine architecture %v not yet implemented", file.FileHeader.Machine)
	}

	// base offset of all relative code addrs
	// BaseOfData - Size - Offset
	var sectionOffset uint64
	opt, ok := file.OptionalHeader.(*pe.OptionalHeader32)
	if !ok {
		return nil, fmt.Errorf("support for optional header type %T not yet implemented", opt)
	}

	// rdata offset range
	rdata := new(RdataRange)

	var section *pe.Section
	for _, sec := range file.Sections {
		if sec.Name == ".text" {
			section = sec
		}

		if sec.Name == ".rdata" {
			rdata.StartOffset = uint64(sec.Offset)
			rdata.EndOffset = rdata.StartOffset + uint64(sec.Size)
		}
	}
	if section == nil {
		return nil, fmt.Errorf("could not find text section")
	}
	sectionOffset += uint64(opt.ImageBase)
	sectionOffset += uint64(opt.BaseOfData)
	sectionOffset -= uint64(section.Size)
	sectionOffset -= uint64(section.Offset)

	s := Section{section, arch, sectionOffset, rdata}
	return &s, nil
}

func checkValidSeq(inst *x86asm.Inst, prev *x86asm.Inst, base uint64, rdata *RdataRange) (uint64, bool) {
	memCurrent, typeOk := inst.Args[0].(x86asm.Mem)
	if !typeOk {
		return 0, false
	}
	immCurrent, typeOk := inst.Args[1].(x86asm.Imm)
	if !typeOk {
		return 0, false
	}

	memPrevious, typeOk := prev.Args[0].(x86asm.Mem)
	if !typeOk {
		return 0, false
	}
	immPrevious, typeOk := prev.Args[1].(x86asm.Imm)
	if !typeOk {
		return 0, false
	}

	// the C code looks something like this
	//   NvFBCCreateParams createParams;
	//   memset(&createParams, 0, sizeof(createParams));
	//   ...
	//   createParams.pPrivateData = (void*)enableKey;
	//   createParams.dwPrivateDataSize = 16;

	// so we want to find an asignment (MOV dword ptr) to an offset (ptr) in .rdata
	// followed by an assignment (MOV dword ptr) of value 16 (10h)

	relAddr := uint64(immPrevious) - base
	diff := memCurrent.Disp - memPrevious.Disp
	if inst.Opcode == 0xc7850000 && (relAddr >= rdata.StartOffset) && (relAddr <= rdata.EndOffset) && immCurrent == 0x10 && diff == 4 {
		return relAddr, true
	}

	return 0, false
}

func getOffset(buf []byte) (uint64, error) {
	sect, err := getTextSection(buf)
	if err != nil {
		return 0, err
	}
	sec := sect.section
	base := sect.base
	rdata := sect.rdata

	raw, err := sec.Data()
	if err != nil {
		return 0, err
	}
	data := raw
	fileSize := len(raw)
	memSize := int(sec.VirtualSize)
	if fileSize > memSize {
		// Ignore section alignment padding.
		data = raw[:memSize]
	}

	code := data
	start := uint64(0)
	end := uint64(len(code))

	var prev *x86asm.Inst
	for pc := start; pc < end; {
		addr := pc

		inst, err := x86asm.Decode(code[addr:], sect.arch)
		size := inst.Len

		pc += uint64(size)

		// handle instruction decode failures
		if err != nil {
			continue
		}

		// let's decode it
		if prev != nil {
			if offset, ok := checkValidSeq(&inst, prev, base, rdata); ok {
				// // debug stuff
				// fmt.Println("got ya ========================")
				// fmt.Printf("%# v\n", pretty.Formatter(prev))
				// fmt.Printf("%# v\n", pretty.Formatter(inst))
				return offset, nil
			}
		}

		prev = &inst
	}

	return 0, fmt.Errorf("could not find offset in DLL")
}

func getPrivData() ([]byte, error) {
	var buf []byte
	if _, err := os.Stat(steamPath); err == nil {
		buf, err = ioutil.ReadFile(steamPath)
		if err != nil {
			return nil, err
		}
	} else if os.IsNotExist(err) {
		buf, err = getFromServer()
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("Could not retrive data from disk or server")
	}

	offset, err := getOffset(buf)
	if err != nil {
		return nil, err
	}

	data := buf[offset : offset+16]
	return data, nil
}

func checkValidData(priv []byte) bool {
	sum := crc32.ChecksumIEEE(priv)
	if sum != validChecksum {
		return false
	}
	return true
}
