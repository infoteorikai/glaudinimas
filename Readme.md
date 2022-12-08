# Glaudinimas
https://lt.wikipedia.org/wiki/Glaudinimas

### Usage:
```
go build ./src/lzw
go build ./src/unlzw
./lzw -help
./lzw -in in.txt -out out.txt
./unlzw -in out.txt -out de.txt

(lzw.exe or unlzw.exe on windows)
```

### TODO
#### LZW
 - [ ] make reset work
 - [ ] use dynamic record sizing
 - [ ] add tests
 - [ ] document code
 - [ ] analyze the compress ratio
