/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tap

// singer-tap specific types
type (
	// SingerTapSchema the structure of this is determined by the catalog/properties JSON of a singer tap
	SingerTapSchema map[string]interface{}
	// SingerTapMetadata the structure of this is determined by the catalog/properties JSON of a singer tap
	SingerTapMetadata map[string]interface{}
	// SingerTapStream the deserialized version of each stream entry in the catalog/properties JSON of a singer tap
	SingerTapStream struct {
		Stream        string              `json:"stream"`
		TapStreamId   string              `json:"tap_stream_id"`
		Schema        SingerTapSchema     `json:"schema"`
		Metadata      []SingerTapMetadata `json:"metadata"`
		KeyProperties interface{}         `json:"key_properties"`
	}
	// SingerTapConfig the set of variables needed to initialize a SingerTapImpl
	SingerTapConfig struct {
		Config               interface{}
		Cmd                  string
		StreamPropertiesFile string
		TapSchemaSetter      func(stream *SingerTapStream)
	}

	// SingerTapProperties wraps SingerTapStreams
	SingerTapProperties struct {
		Streams []*SingerTapStream `json:"streams"`
	}
)
