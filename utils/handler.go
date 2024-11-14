// Copyright 2024-2025 NetCracker Technology Corporation
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

package utils

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
)

type CustomLogHandler struct {
	slog.Handler
	l *log.Logger
}

func (h *CustomLogHandler) Handle(ctx context.Context, record slog.Record) error {
	level := fmt.Sprintf("[%v]", record.Level.String())
	timeStr := record.Time.Format("[2006-01-02T15:04:05.999]")
	msg := record.Message

	h.l.Println(timeStr, level, "[request_id= ]", "[tenant_id= ]", "[thread= ]", "[class= ]", msg)

	return nil
}

func NewCustomLogHandler(out io.Writer) *CustomLogHandler {
	h := &CustomLogHandler{
		Handler: slog.NewTextHandler(out, nil),
		l:       log.New(out, "", 0),
	}
	return h
}
