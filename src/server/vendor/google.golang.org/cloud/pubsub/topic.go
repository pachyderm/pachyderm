// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pubsub

import (
	"fmt"

	"golang.org/x/net/context"
)

const MaxPublishBatchSize = 1000

// TopicHandle is a reference to a PubSub topic.
type TopicHandle struct {
	c *Client

	// The fully qualified identifier for the topic, in the format "projects/<projid>/topics/<name>"
	name string
}

// NewTopic creates a new topic.
// The specified topic name must start with a letter, and contain only letters
// ([A-Za-z]), numbers ([0-9]), dashes (-), underscores (_), periods (.),
// tildes (~), plus (+) or percent signs (%). It must be between 3 and 255
// characters in length, and must not start with "goog".
// If the topic already exists an error will be returned.
func (c *Client) NewTopic(ctx context.Context, name string) (*TopicHandle, error) {
	t := c.Topic(name)
	err := c.s.createTopic(ctx, t.Name())
	return t, err
}

// Topic creates a reference to a topic.
func (c *Client) Topic(name string) *TopicHandle {
	return &TopicHandle{c: c, name: fmt.Sprintf("projects/%s/topics/%s", c.projectID, name)}
}

// Topics lists all of the topics for the client's project.
func (c *Client) Topics(ctx context.Context) ([]*TopicHandle, error) {
	topicNames, err := c.s.listProjectTopics(ctx, c.fullyQualifiedProjectName())
	if err != nil {
		return nil, err
	}

	topics := []*TopicHandle{}
	for _, t := range topicNames {
		topics = append(topics, &TopicHandle{c: c, name: t})
	}
	return topics, nil
}

// Name returns the globally unique name for the topic.
func (t *TopicHandle) Name() string {
	return t.name
}

// Delete deletes the topic.
func (t *TopicHandle) Delete(ctx context.Context) error {
	return t.c.s.deleteTopic(ctx, t.name)
}

// Exists reports whether the topic exists on the server.
func (t *TopicHandle) Exists(ctx context.Context) (bool, error) {
	if t.name == "_deleted-topic_" {
		return false, nil
	}

	return t.c.s.topicExists(ctx, t.name)
}

// Subscriptions lists the subscriptions for this topic.
func (t *TopicHandle) Subscriptions(ctx context.Context) ([]*SubscriptionHandle, error) {
	subNames, err := t.c.s.listTopicSubscriptions(ctx, t.name)
	if err != nil {
		return nil, err
	}

	subs := []*SubscriptionHandle{}
	for _, s := range subNames {
		subs = append(subs, &SubscriptionHandle{c: t.c, name: s})
	}
	return subs, nil
}

// Publish publishes the supplied Messages to the topic.
// If successful, the server-assigned message IDs are returned in the same order as the supplied Messages.
// At most MaxPublishBatchSize messages may be supplied.
func (t *TopicHandle) Publish(ctx context.Context, msgs ...*Message) ([]string, error) {
	if len(msgs) == 0 {
		return nil, nil
	}
	if len(msgs) > MaxPublishBatchSize {
		return nil, fmt.Errorf("pubsub: got %d messages, but maximum batch size is %d", len(msgs), MaxPublishBatchSize)
	}
	return t.c.s.publishMessages(ctx, t.name, msgs)
}
