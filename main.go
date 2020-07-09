// Copyright 2013 Prometheus Team
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

package main

import (
	"log"
)

func main() {
	log.Print("Starting event suppressor...")
	/*
		镇压者包含内容(相当于后来的静默)：
		1. Suppressions, 镇压内容, 多个
			1. id
			2. Description, 描述
			3. Filters, 过滤器(k-v, fingerprint)
			4. EndsAt
			5. CreatedBy
			6. CreatedAt
		2. suppressionReqs, 镇压请求
		3. suppressionSummaryReqs, 镇压小结请求
		4. isInhibitedReqs
	*/
	suppressor := NewSuppressor()
	defer suppressor.Close()
	/*
		dispatch 逻辑
		1. 如果有 suppressionReqs 则 dispatchSuppression
			1. 堆也叫优先队列, list 实现即可. 将 Suppression 入队
			2. suppressionRequest.Response 初始化
		2. 如果 isInhibitedReqs 则 queryInhibit
			1. queryInhibit
		3. 如果 suppressionSummaryReqs 则 generateSummary
			1. generateSummary
		4. 每隔 30s 进行一次 reapSuppressions
			1. 取出当前时间后边 Suppressions
			2. 根据当前的 Suppressions 初始化优先队列, 上调或下沉
	*/
	go suppressor.Dispatch()
	log.Println("Done.")

	log.Println("Starting event aggregator...")
	aggregator := NewAggregator()
	defer aggregator.Close()

	summarizer := new(SummaryDispatcher)
	go aggregator.Dispatch(summarizer)
	log.Println("Done.")

	done := make(chan bool)
	go func() {
		rules := AggregationRules{
			&AggregationRule{
				Filters: Filters{NewFilter("service", "discovery")},
			},
		}

		aggregator.SetRules(rules)

		events := Events{
			&Event{
				Payload: map[string]string{
					"service": "discovery",
				},
			},
		}

		aggregator.Receive(events)

		done <- true
	}()
	<-done

	log.Println("Running summary dispatcher...")
	summarizer.Dispatch(suppressor)
}
