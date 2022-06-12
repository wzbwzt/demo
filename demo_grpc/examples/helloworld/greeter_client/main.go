/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package main implements a client for Greeter service.
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/opentracing/opentracing-go"
	"github.com/openzipkin/zipkin-go"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"

	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
)

const (
	defaultName                = "Test"
	zipkin_report_endpoint_url = "http://10.4.7.140:9411/api/v2/spans"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
	name = flag.String("name", defaultName, "Name to greet")
)

func main() {
	flag.Parse()

	// Create a metrics registry.
	reg := prometheus.NewRegistry()
	// Create some standard client metrics.
	grpcMetrics := grpc_prometheus.NewClientMetrics()
	// Register client metrics to registry.
	reg.MustRegister(grpcMetrics)

	//newtrace
	tracer, err := newtrace("10.4.7.140:5051")
	if err != nil {
		log.Fatalf("创建trace失败,err:%s", err)
	}

	// Set up a connection to the server.
	conn, err := grpc.Dial(
		*addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(
			grpc_middleware.ChainUnaryClient(
				grpcMetrics.UnaryClientInterceptor(),
				otgrpc.OpenTracingClientInterceptor(tracer, otgrpc.LogPayloads())),
		),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.
	go func() {
		for {
			r, err := c.SayHello(context.Background(), &pb.HelloRequest{Name: *name})
			if err != nil {
				log.Fatalf("could not greet: %v", err)
			}
			log.Printf("Greeting: %s", r.GetMessage())

			time.Sleep(time.Second * 3)
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("You can press n or N to stop the process of client")
	for scanner.Scan() {
		if strings.ToLower(scanner.Text()) == "n" {
			os.Exit(0)
		}
	}
}

func newtrace(localhost string) (tracer opentracing.Tracer, err error) {
	// set up a span reporter
	reporter := zipkinhttp.NewReporter(zipkin_report_endpoint_url)
	// defer reporter.Close()

	// create our local service endpoint
	endpoint, err := zipkin.NewEndpoint("test_client", localhost)
	if err != nil {
		return
	}

	// initialize our tracer
	nativeTracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(endpoint))
	if err != nil {
		return
	}

	// use zipkin-go-opentracing to wrap our tracer
	tracer = zipkinot.Wrap(nativeTracer)

	// optionally set as Global OpenTracing tracer instance
	opentracing.SetGlobalTracer(tracer)
	return
}
