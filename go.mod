module github.com/InsideGallery/core

go 1.24.0

retract (
	v1.0.2 // Remove miss-use of mongo config hosts
	v1.0.1 // Version contains broken tests
)

require (
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/gops v0.3.28
	github.com/jmoiron/sqlx v1.3.5
	github.com/nats-io/nats.go v1.34.1
	github.com/pkg/errors v0.9.1
	github.com/tidwall/buntdb v1.3.0
	github.com/valyala/fasthttp v1.65.0
	go.mongodb.org/mongo-driver v1.15.0
	go.uber.org/atomic v1.11.0
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	google.golang.org/grpc v1.65.0 // indirect
)

require (
	dario.cat/mergo v1.0.1
	github.com/AlekSi/pointer v1.2.0
	github.com/DataDog/datadog-api-client-go/v2 v2.44.0
	github.com/Masterminds/semver/v3 v3.4.0
	github.com/aerospike/aerospike-client-go/v7 v7.2.1
	github.com/agoda-com/opentelemetry-go/otelslog v0.1.1
	github.com/agoda-com/opentelemetry-logs-go v0.5.0
	github.com/apache/tinkerpop/gremlin-go/v3 v3.7.2
	github.com/caarlos0/env/v10 v10.0.0
	github.com/dgryski/go-farm v0.0.0-20240924180020-3414d57e47da
	github.com/dgryski/go-minhash v0.0.0-20190315135803-ad340ca03076
	github.com/dgryski/go-spooky v0.0.0-20170606183049-ed3d087f40e2
	github.com/elastic/go-elasticsearch/v8 v8.16.0
	github.com/go-faster/xor v1.0.0
	github.com/go-jose/go-jose/v3 v3.0.4
	github.com/go-redsync/redsync/v4 v4.13.0
	github.com/go-slog/otelslog v0.1.0
	github.com/gofiber/fiber/v2 v2.52.5
	github.com/gofiber/jwt/v4 v4.0.0
	github.com/golang-jwt/jwt/v5 v5.2.2
	github.com/golang/mock v1.6.0
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.5.5
	github.com/mailru/easyjson v0.7.7
	github.com/mfonda/simhash v0.0.0-20151007195837-79f94a1100d6
	github.com/nats-io/nkeys v0.4.7
	github.com/neo4j/neo4j-go-driver/v5 v5.20.0
	github.com/redis/go-redis/v9 v9.5.5
	github.com/samber/lo v1.47.0
	github.com/samber/slog-datadog/v2 v2.8.0
	github.com/samber/slog-multi v1.0.3
	github.com/samber/slog-otel v0.1.0
	github.com/segmentio/go-hll v1.0.1
	github.com/sirbu/golang-common v0.0.0-20170403140351-21d4febd4bca
	github.com/spf13/cast v1.6.0
	github.com/stretchr/testify v1.11.1
	github.com/sugarme/tokenizer v0.3.0
	github.com/tink-crypto/tink-go/v2 v2.5.0
	github.com/twmb/murmur3 v1.1.8
	go.opentelemetry.io/contrib/instrumentation/host v0.52.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.53.0
	go.opentelemetry.io/contrib/instrumentation/runtime v0.52.0
	go.opentelemetry.io/otel v1.28.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.27.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.27.0
	go.opentelemetry.io/otel/metric v1.28.0
	go.opentelemetry.io/otel/sdk v1.27.0
	go.opentelemetry.io/otel/sdk/metric v1.27.0
	go.opentelemetry.io/otel/trace v1.28.0
	go.uber.org/mock v0.5.0
	golang.org/x/crypto v0.41.0
	golang.org/x/text v0.28.0
	gorgonia.org/gorgonia v0.9.18
	gorgonia.org/tensor v0.9.24
)

require (
	github.com/DataDog/zstd v1.5.2 // indirect
	github.com/MicahParks/keyfunc/v2 v2.0.3 // indirect
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/apache/arrow/go/arrow v0.0.0-20211112161151-bc219186db40 // indirect
	github.com/awalterschulze/gographviz v2.0.3+incompatible // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/chewxy/hm v1.0.0 // indirect
	github.com/chewxy/math32 v1.10.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-metro v0.0.0-20250106013310-edb8663e5e33 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/elastic/elastic-transport-go/v8 v8.6.0 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/gomodule/redigo v2.0.0+incompatible // indirect
	github.com/google/flatbuffers v2.0.8+incompatible // indirect
	github.com/gorilla/websocket v1.5.1 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.20.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/leesper/go_rng v0.0.0-20190531154944-a612b043e353 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/lufia/plan9stats v0.0.0-20240513124658-fba389f38bae // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/mattn/go-sqlite3 v1.14.16 // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/montanaflynn/stats v0.6.6 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/nicksnyder/go-i18n/v2 v2.4.0 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pierrec/lz4/v4 v4.1.16 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/samber/slog-common v0.17.0 // indirect
	github.com/schollz/progressbar/v2 v2.15.0 // indirect
	github.com/shirou/gopsutil/v3 v3.24.4 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/sugarme/regexpset v0.0.0-20200920021344-4d4ec8eaf93c // indirect
	github.com/tidwall/btree v1.4.2 // indirect
	github.com/tidwall/gjson v1.14.3 // indirect
	github.com/tidwall/grect v0.1.4 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/tidwall/rtred v0.1.2 // indirect
	github.com/tidwall/tinyqueue v0.1.1 // indirect
	github.com/tklauser/go-sysconf v0.3.14 // indirect
	github.com/tklauser/numcpus v0.8.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/xtgo/set v1.0.0 // indirect
	github.com/youmark/pkcs8 v0.0.0-20181117223130-1be2e3e5546d // indirect
	github.com/yuin/gopher-lua v1.1.1 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.27.0 // indirect
	go.opentelemetry.io/proto/otlp v1.2.0 // indirect
	go4.org/unsafe/assume-no-moving-gc v0.0.0-20220617031537-928513b29760 // indirect
	golang.org/x/exp v0.0.0-20240525044651-4c93da0ed11d // indirect
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/oauth2 v0.20.0 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	gonum.org/v1/gonum v0.11.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240528184218-531527333157 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240528184218-531527333157 // indirect
	google.golang.org/protobuf v1.36.8 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gorgonia.org/cu v0.9.4 // indirect
	gorgonia.org/dawson v1.2.0 // indirect
	gorgonia.org/vecf32 v0.9.0 // indirect
	gorgonia.org/vecf64 v0.9.0 // indirect
)
