#!/usr/bin/env bash
#
# Developper script to build binaries
#
set -o nounset # Treat unset variables as an error

OUTPUT_DIR="./build"

BINARY_ARMV7_NAME="gofetchmyfeeds-ARMv7"
BINARY_X86_NAME="gofetchmyfeeds-AMD64"

PACKAGE_ARCHIVE_NAME="gofetchmyfeeds.zip"

BUILD_ARMV7_CMD="env GOOS=linux GOARCH=arm GOARM=7 go build -o $OUTPUT_DIR/$BINARY_ARMV7_NAME ."
BUILD_X86_CMD="go build -o $OUTPUT_DIR/$BINARY_X86_NAME ."

echo "-cleaning $OUTPUT_DIR"
rm -Rf $OUTPUT_DIR/*

echo "-building $BINARY_X86_NAME"
$BUILD_X86_CMD

echo "-building $BINARY_ARMV7_NAME"
$BUILD_ARMV7_CMD

echo "-building binaries zip"
zip -r $OUTPUT_DIR/$PACKAGE_ARCHIVE_NAME $OUTPUT_DIR
