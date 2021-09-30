// Copyright 2015 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build !nomeminfo
// +build !nomeminfo

package collector

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	reParens = regexp.MustCompile(`\((.*)\)`)
)

func (c *meminfoCollector) getMemInfo() (map[string]float64, error) {
	file, err := os.Open(procFilePath("meminfo"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return parseMemInfo(file)
}

func parseMemInfo(r io.Reader) (map[string]float64, error) {
	var (
		memInfo = map[string]float64{}
		scanner = bufio.NewScanner(r)
	)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		// Workaround for empty lines occasionally occur in CentOS 6.2 kernel 3.10.90.
		if len(parts) == 0 {
			continue
		}
		if len(parts) < 2 || len(parts) > 3 {
			return nil, fmt.Errorf("invalid line in meminfo: %s", line)
		}

		fv, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid value in meminfo: %w", err)
		}
		key := parts[0][:len(parts[0])-1] // remove trailing : from key
		// Active(anon) -> Active_anon
		key = reParens.ReplaceAllString(key, "_${1}")

		// unitless values are assumed to be in bytes; unitted values
		// are presumed to be kB.
		if len(parts) == 3 {
			fv *= 1024
			key = key + "_bytes"
		}

		memInfo[key] = fv
	}

	return memInfo, scanner.Err()
}
