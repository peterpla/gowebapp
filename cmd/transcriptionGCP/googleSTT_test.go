package main

import (
	"testing"

	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"
)

func TestProcessTranscription(t *testing.T) {

	type test struct {
		name      string
		wordElems []*speechpb.WordInfo
		result    []string
	}

	tests := []test{
		{"one word",
			[]*speechpb.WordInfo{
				{Word: "Thank", SpeakerTag: 1},
			},
			[]string{
				"[Speaker 1] Thank\n",
			},
		},

		{"two words",
			[]*speechpb.WordInfo{
				{Word: "Thank", SpeakerTag: 1},
				{Word: "you", SpeakerTag: 1},
			},
			[]string{
				"[Speaker 1] Thank you\n",
			},
		},

		{"two speakers",
			[]*speechpb.WordInfo{
				{Word: "Thank", SpeakerTag: 1},
				{Word: "you", SpeakerTag: 1},
				{Word: "for", SpeakerTag: 1},
				{Word: "calling", SpeakerTag: 1},
				{Word: "Park", SpeakerTag: 1},
				{Word: "flooring.", SpeakerTag: 1},
				{Word: "This", SpeakerTag: 1},
				{Word: "is", SpeakerTag: 1},
				{Word: "Michael.", SpeakerTag: 1},
				{Word: "How", SpeakerTag: 1},
				{Word: "may", SpeakerTag: 1},
				{Word: "I", SpeakerTag: 1},
				{Word: "help", SpeakerTag: 1},
				{Word: "you?", SpeakerTag: 1},
				{Word: "Hey", SpeakerTag: 2},
				{Word: "Michael.", SpeakerTag: 2},
				{Word: "How", SpeakerTag: 2},
				{Word: "are", SpeakerTag: 2},
				{Word: "you", SpeakerTag: 2},
				{Word: "today?", SpeakerTag: 2},
				{Word: "Good.", SpeakerTag: 1},
				{Word: "What's", SpeakerTag: 1},
				{Word: "up?", SpeakerTag: 1},
				{Word: "My", SpeakerTag: 2},
				{Word: "name", SpeakerTag: 2},
				{Word: "is", SpeakerTag: 2},
				{Word: "Yuri.", SpeakerTag: 2},
			},
			[]string{
				"[Speaker 1] Thank you for calling Park flooring. This is Michael. How may I help you?|",
				"[Speaker 2] Hey Michael. How are you today?|",
				"[Speaker 1] Good. What's up?|",
				"[Speaker 2] My name is Yuri.\n",
			},
		},
	}

	for _, tc := range tests {

		returnSlice := wordsToAttributedStrings(tc.wordElems)

		for i, returnString := range returnSlice {
			// log.Printf("returnSlice: %+v, wordElems: %+v\n", returnSlice, tc.wordElems)
			if returnString != tc.result[i] {
				t.Errorf("%s: expected %v, got %v", tc.name, tc.result[i], returnSlice)
			}
		}
	}
}
