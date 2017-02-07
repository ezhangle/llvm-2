#!/bin/bash
if [ $# -lt 1 ]; then
	echo "Usage: strip FILE.ll"
	exit 1
fi
f=$1
# Replace basic block label name comments with labels.
sar -i "; <label>:([0-9]+):[^\n]+" "\${1}:" "${f}"
# Remove comments.
sar -i "(^|[\n]);[^\n]+" "" "${f}"
# Remove target specifiers.
sar -i "(^|[\n])target[^\n]+" "" "${f}"
# Remove attributes.
sar -i "(^|[\n])attributes[^\n]+" "" "${f}"
# Remove function attributes.
sar -i "[)] #[0-9]+" ")" "${f}"
# Remove metadata nodes.
sar -i "(^|[\n])[!][^\n]+" "" "${f}"
# Remove alignment.
sar -i ", align [0-9]+\n" "\n" "${f}"
# Add labels for the first basic block of functions with 0 parameters.

# TODO: Fix handling of ...
# e.g. define void @sqlite3VdbeMultiLoad(%struct.Vdbe*, i32, i8*, ...) {
# should be 3, got 4.
sar -i "([\n]define[^@]+[@][^(]+[(]([^(]+[(][^)]+[)][*]|[^),]+)(, ([^(]+[(][^)]+[)][*]|[^),]+)){14}[)] {)\n([^0-9])" "\${1}\n15:\n\${5}" "${f}"
sar -i "([\n]define[^@]+[@][^(]+[(]([^(]+[(][^)]+[)][*]|[^),]+)(, ([^(]+[(][^)]+[)][*]|[^),]+)){13}[)] {)\n([^0-9])" "\${1}\n14:\n\${5}" "${f}"
sar -i "([\n]define[^@]+[@][^(]+[(]([^(]+[(][^)]+[)][*]|[^),]+)(, ([^(]+[(][^)]+[)][*]|[^),]+)){12}[)] {)\n([^0-9])" "\${1}\n13:\n\${5}" "${f}"
sar -i "([\n]define[^@]+[@][^(]+[(]([^(]+[(][^)]+[)][*]|[^),]+)(, ([^(]+[(][^)]+[)][*]|[^),]+)){11}[)] {)\n([^0-9])" "\${1}\n12:\n\${5}" "${f}"
sar -i "([\n]define[^@]+[@][^(]+[(]([^(]+[(][^)]+[)][*]|[^),]+)(, ([^(]+[(][^)]+[)][*]|[^),]+)){10}[)] {)\n([^0-9])" "\${1}\n11:\n\${5}" "${f}"
sar -i "([\n]define[^@]+[@][^(]+[(]([^(]+[(][^)]+[)][*]|[^),]+)(, ([^(]+[(][^)]+[)][*]|[^),]+)){9}[)] {)\n([^0-9])" "\${1}\n110:\n\${5}" "${f}"
sar -i "([\n]define[^@]+[@][^(]+[(]([^(]+[(][^)]+[)][*]|[^),]+)(, ([^(]+[(][^)]+[)][*]|[^),]+)){8}[)] {)\n([^0-9])" "\${1}\n9:\n\${5}" "${f}"
sar -i "([\n]define[^@]+[@][^(]+[(]([^(]+[(][^)]+[)][*]|[^),]+)(, ([^(]+[(][^)]+[)][*]|[^),]+)){7}[)] {)\n([^0-9])" "\${1}\n8:\n\${5}" "${f}"
sar -i "([\n]define[^@]+[@][^(]+[(]([^(]+[(][^)]+[)][*]|[^),]+)(, ([^(]+[(][^)]+[)][*]|[^),]+)){6}[)] {)\n([^0-9])" "\${1}\n7:\n\${5}" "${f}"
sar -i "([\n]define[^@]+[@][^(]+[(]([^(]+[(][^)]+[)][*]|[^),]+)(, ([^(]+[(][^)]+[)][*]|[^),]+)){5}[)] {)\n([^0-9])" "\${1}\n6:\n\${5}" "${f}"
sar -i "([\n]define[^@]+[@][^(]+[(]([^(]+[(][^)]+[)][*]|[^),]+)(, ([^(]+[(][^)]+[)][*]|[^),]+)){4}[)] {)\n([^0-9])" "\${1}\n5:\n\${5}" "${f}"
sar -i "([\n]define[^@]+[@][^(]+[(]([^(]+[(][^)]+[)][*]|[^),]+)(, ([^(]+[(][^)]+[)][*]|[^),]+)){3}[)] {)\n([^0-9])" "\${1}\n4:\n\${5}" "${f}"
sar -i "([\n]define[^@]+[@][^(]+[(]([^(]+[(][^)]+[)][*]|[^),]+)(, ([^(]+[(][^)]+[)][*]|[^),]+)){2}[)] {)\n([^0-9])" "\${1}\n3:\n\${5}" "${f}"
sar -i "([\n]define[^@]+[@][^(]+[(]([^(]+[(][^)]+[)][*]|[^),]+)(, ([^(]+[(][^)]+[)][*]|[^),]+)){1}[)] {)\n([^0-9])" "\${1}\n2:\n\${5}" "${f}"
sar -i "([\n]define[^@]+[@][^(]+[(]([^(]+[(][^)]+[)][*]|[^),]+)(, ([^(]+[(][^)]+[)][*]|[^),]+)){0}[)] {)\n([^0-9])" "\${1}\n1:\n\${5}" "${f}"
sar -i "([\n]define[^@]+[@][^(]+[(]([^(]+[(][^)]+[)][*]|[^),]+)(, ([^(]+[(][^)]+[)][*]|[^),]+)){0}[)] {)\n([^0-9])" "\${1}\n0:\n\${5}" "${f}"
sar -i "([\n]define[^@]+[@][^(]+[(][)] {)\n([^0-9])" "\${1}\n0:\n\${2}" "${f}"

sar -i "\n\n" "\n" "${f}"
sar -i "\n\n" "\n" "${f}"
sar -i "^\n" "" "${f}"
sar -i "[\n]  " "\n\t" "${f}"
sar -i "[\n][\t]  " "\n\t\t" "${f}"

sar -i "([0-9]+):\n" "; <label>:\${1}\n" "${f}"

sar -i "getelementptr inbounds " "getelementptr " "${f}"
sar -i "= common " "= " "${f}"
sar -i "unnamed_addr constant " "constant " "${f}"
sar -i "[\n]define internal " "\ndefine " "${f}"
sar -i "internal (constant|global) " "\${1} " "${f}"
sar -i "(i8|i16|i32|call|define) signext" "\${1}" "${f}"
sar -i "(i8|i16|i32|call|define) zeroext" "\${1}" "${f}"
sar -i " noalias " " " "${f}"
sar -i "[*] nocapture" "*" "${f}"
sar -i "[*] writeonly" "*" "${f}"
sar -i "[*] readonly" "*" "${f}"
sar -i "[*] nocapture" "*" "${f}"
sar -i " volatile " " " "${f}"
#sar -i "getelementptr ([^,]+), ([^,]+), i32 0, i32 0" "getelementptr \${1}, \${2}, i64 0, i64 0" "${f}"
sar -i "[.]000000e[+]00" ".0" "${f}"
sar -i "([0-9])[.]([0-9]+)?[0]+e[+]([0-9][0-9])" "\${1}.\${2}e\${3}" "${f}"
sar -i "([0-9])[.]000000e-01" "0.\${1}" "${f}"
