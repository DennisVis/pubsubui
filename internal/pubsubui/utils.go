// Copyright 2022 Dennis Vis
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pubsubui

import (
	"log"
	"os"
)

func logWithPrefix(format string, a ...interface{}) {
	log.Printf(LogPrefix+format, a...)
}

func envOrDefault(key, def string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return def
	}

	return val
}

func filterEmptyStrings(strs []string) []string {
	filtered := make([]string, 0)

	for _, s := range strs {
		if s != "" {
			filtered = append(filtered, s)
		}
	}

	return filtered
}

func deduplicateStrings(strs []string) []string {
	stringsMap := make(map[string]bool)

	for _, s := range strs {
		stringsMap[s] = true
	}

	strings := make([]string, len(stringsMap))
	idx := 0
	for s := range stringsMap {
		strings[idx] = s
		idx++
	}

	return strings
}
