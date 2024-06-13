package goqueue_test

// func TestQueueService_Start(t *testing.T) {
// 	// Create a mock consumer
// 	mockConsumer := &mocks.Consumer{}

// 	// Create a mock handler
// 	mockHandler := &mocks.InboundMessageHandler{}

// 	// Create a mock publisher
// 	mockPublisher := &mocks.Publisher{}

// 	// Create a new QueueService instance with the mock objects
// 	qs := goqueue.NewQueueService(
// 		goqueue.WithConsumer(mockConsumer),
// 		goqueue.WithMessageHandler(mockHandler),
// 		goqueue.WithPublisher(mockPublisher),
// 		goqueue.WithNumberOfConsumer(3),
// 	)

// 	// Set up an expectation for the Consume method
// 	mockConsumer.On("Consume", mock.Anything, mock.Anything, mock.Anything).
// 		Return(func(ctx context.Context, handler goqueue.InboundMessageHandler, meta map[string]interface{}) (err error) {
// 			time.Sleep(time.Second) // simulate a message being consumed
// 			return nil
// 		}).Times(qs.NumberOfConsumer) //total worker spawned

// 	// Create a context for the test
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
// 	defer cancel()
// 	// Start the QueueService
// 	go func() {
// 		err := qs.Start(ctx)
// 		require.NoError(t, err)
// 		cancel() // cancel the context to stop the QueueService
// 	}()
// 	<-ctx.Done()
// 	mockConsumer.AssertCalled(t, "Consume", mock.Anything, mock.Anything, mock.Anything)
// }

// func TestQueueService_Publish(t *testing.T) {
// 	// Create a mock publisher
// 	mockPublisher := &mocks.Publisher{}
// 	// Create a new QueueService instance with the mock publisher
// 	qs := goqueue.NewQueueService(
// 		goqueue.WithPublisher(mockPublisher),
// 	)

// 	// Create a test message
// 	message := goqueue.Message{
// 		ID:   "1",
// 		Data: map[string]interface{}{"key": "value"},
// 	}

// 	// Set up an expectation for the Publish method
// 	mockPublisher.On("Publish", mock.Anything, message).
// 		Return(nil)

// 	// Call the Publish method
// 	err := qs.Publish(context.Background(), message)

// 	// Assert that no error occurred
// 	require.NoError(t, err)

// 	// Assert that the Publish method was called with the correct arguments
// 	mockPublisher.AssertCalled(t, "Publish", mock.Anything, message)
// }
