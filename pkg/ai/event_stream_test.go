package ai

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestEventStream(t *testing.T) {
	// 测试代码
	stream := NewEventStream[AssistantMessageEvent, any](
		func(event AssistantMessageEvent) bool {
			return event.(*AssistantMessageEventDone) != nil || event.(*AssistantMessageEventError) != nil
		},
		func(event AssistantMessageEvent) any {
			if event.(*AssistantMessageEventDone) != nil {
				return event.(*AssistantMessageEventDone).Message
			} else if event.(*AssistantMessageEventError) != nil {
				return event.(*AssistantMessageEventError).Error
			}
			panic("Unexpected event type for final result")
		},
	)

	go func() {
		stream.Push(NewAssistantMessageEventStart(nil))
		stream.Push(NewAssistantMessageEventTextStart(0, nil))
		stream.Push(NewAssistantMessageEventTextDelta(0, "Hi", nil))
		stream.Push(NewAssistantMessageEventTextEnd(0, "Hi", nil))
		stream.Push(NewAssistantMessageEventDone(StopReasonStop, &AssistantMessage{
			Role:       MessageRoleAssistant,
			StopReason: StopReasonStop,
		}))

	}()

	// 使用通道
	for event := range stream.Events() {
		jsonEvent, _ := json.Marshal(event)
		fmt.Println(string(jsonEvent))
	}

	result := stream.Result()
	jsonResult, _ := json.Marshal(result)
	fmt.Println(string(jsonResult))
}
