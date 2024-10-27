# mus-stream-dts-go
mus-stream-dts-go provides DTM (Data Type Metadata) support for the 
[mus-stream-go](https://github.com/mus-format/mus-stream-go) serializer. With 
mus-stream-dts-go we can encode/decode DTM + data itself.

It completely repeats the structure of [mus-dts-go](https://github.com/mus-format/mus-dts-go), 
and differs only in that it uses `Writer`, `Reader` interfaces rather than а 
slice of bytes.

# Tests
Test coverage is 100%.

# How To Use
You can learn more about this in the mus-dts-go [documentation](https://github.com/mus-format/mus-dts-go).