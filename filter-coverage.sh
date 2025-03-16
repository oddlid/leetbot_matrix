# Pipe the output of go_cover to this to only see stuff not fully covered
gawk '{if ($3 != "100.0%") print $0}'
