# Gdenticon

This is a translation from the JavaScript implementation [Jdenticon](www.jdenticon.com) to the Go language. It uses the exact same algorithm to produce [Identicons](https://en.wikipedia.org/wiki/Identicon). However, the Gdenticon only accept hashes now (at least 11 chars) and only produce [SVGs](https://en.wikipedia.org/wiki/Scalable_Vector_Graphics) for simplicity.

# How to use?

Type `go build gdenticon.go` to produce a executable, and type `./gdenticon $HASH $OUTPUT_FILE` to use it.
