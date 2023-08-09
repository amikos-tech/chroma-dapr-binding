package chroma_dapr

import (
	"context"
	"fmt"
	"github.com/dapr/components-contrib/bindings"
	"github.com/dapr/components-contrib/metadata"
	"github.com/dapr/kit/logger"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

const (
	testCollection = "testCollection"
)

func loadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Printf("Error loading .env file: %s\n", err)
	}
}

func TestOperations(t *testing.T) {
	t.Parallel()
	t.Run("Get operation list", func(t *testing.T) {
		b := NewChroma(nil)
		assert.NotNil(t, b)
		l := b.Operations()
		assert.Equal(t, 7, len(l))
	})
}

func TestChromaIntegration(t *testing.T) {
	loadEnv()
	url := os.Getenv("CHROMA_URL")
	openAiAPIKey := os.Getenv("OPENAI_API_KEY")
	if url == "" {
		t.SkipNow()
	}

	b := NewChroma(logger.NewLogger("test"))
	m := bindings.Metadata{Base: metadata.Base{Properties: map[string]string{"url": url, "openAIApiKey": openAiAPIKey}}}
	if err := b.Init(context.Background(), m); err != nil {
		t.Fatal(err)
	}

	ctx := context.TODO()
	t.Run("Invoke version", func(t *testing.T) {
		req := &bindings.InvokeRequest{
			Operation: version,
		}
		res, err := b.Invoke(ctx, req)
		assertResponse(t, res, err)
	})

	t.Run("Invoke heartbeat", func(t *testing.T) {
		req := &bindings.InvokeRequest{
			Operation: heartbeat,
		}
		res, err := b.Invoke(ctx, req)
		assertResponse(t, res, err)
	})

	t.Run("Invoke reset", func(t *testing.T) {
		req := &bindings.InvokeRequest{
			Operation: reset,
		}
		res, err := b.Invoke(ctx, req)
		assertResponse(t, res, err)
	})

	t.Run("Invoke list collections", func(t *testing.T) {
		resetReq := &bindings.InvokeRequest{
			Operation: reset,
		}
		resReset, errReset := b.Invoke(ctx, resetReq)
		require.Nil(t, errReset)
		require.NotNil(t, resReset)
		createReq := &bindings.InvokeRequest{
			Operation: createCollection,
			Data: []byte(fmt.Sprintf(`{
				"name": "%s",
				"embeddingFunction": "openai",
				"metadata": {
					"type": "col"
				},	
				"distanceFunction": "l2"
			}`, testCollection)),
		}
		resCreate, createErr := b.Invoke(ctx, createReq)
		require.Nil(t, createErr)
		require.NotNil(t, resCreate)
		req := &bindings.InvokeRequest{
			Operation: listCollections,
		}
		res, err := b.Invoke(ctx, req)
		assertResponse(t, res, err)
	})

	t.Run("Invoke create collection", func(t *testing.T) {
		resetReq := &bindings.InvokeRequest{
			Operation: reset,
		}
		resReset, errReset := b.Invoke(ctx, resetReq)
		require.Nil(t, errReset)
		require.NotNil(t, resReset)
		req := &bindings.InvokeRequest{
			Operation: createCollection,
			Data: []byte(fmt.Sprintf(`{
				"name": "%s",
				"embeddingFunction": "openai",
				"metadata": {
					"type": "col"
				},	
				"distanceFunction": "l2"
			}`, testCollection)),
		}
		res, err := b.Invoke(ctx, req)
		assertResponse(t, res, err)
	})

	t.Run("Invoke delete collection", func(t *testing.T) {
		resetReq := &bindings.InvokeRequest{
			Operation: reset,
		}
		resReset, errReset := b.Invoke(ctx, resetReq)
		require.Nil(t, errReset)
		require.NotNil(t, resReset)
		createReq := &bindings.InvokeRequest{
			Operation: createCollection,
			Data: []byte(fmt.Sprintf(`{
				"name": "%s",
				"embeddingFunction": "openai",
				"metadata": {
					"type": "col"
				},	
				"distanceFunction": "l2"
			}`, testCollection)),
		}
		resCreate, createErr := b.Invoke(ctx, createReq)
		require.Nil(t, createErr)
		require.NotNil(t, resCreate)
		req := &bindings.InvokeRequest{
			Operation: deleteCollection,
			Data: []byte(fmt.Sprintf(`{
				"name": "%s"
			}`, testCollection)),
		}
		res, err := b.Invoke(ctx, req)
		assertResponse(t, res, err)
	})
	t.Run("Invoke delete collection", func(t *testing.T) {
		resetReq := &bindings.InvokeRequest{
			Operation: reset,
		}
		resReset, errReset := b.Invoke(ctx, resetReq)
		require.Nil(t, errReset)
		require.NotNil(t, resReset)
		createReq := &bindings.InvokeRequest{
			Operation: createCollection,
			Data: []byte(fmt.Sprintf(`{
				"name": "%s",
				"embeddingFunction": "openai",
				"metadata": {
					"type": "col"
				},	
				"distanceFunction": "l2"
			}`, testCollection)),
		}
		resCreate, createErr := b.Invoke(ctx, createReq)
		require.Nil(t, createErr)
		require.NotNil(t, resCreate)
		req := &bindings.InvokeRequest{
			Operation: getCollection,
			Data: []byte(fmt.Sprintf(`{
				"name": "%s"
			}`, testCollection)),
		}
		res, err := b.Invoke(ctx, req)
		assertResponse(t, res, err)
	})

}

func assertResponse(t *testing.T, res *bindings.InvokeResponse, err error) {
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Metadata)
	fmt.Printf("Response Meta: %s\n", res.Metadata)
	fmt.Printf("Response: %s\n", res.Data)
}
