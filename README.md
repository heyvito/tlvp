# tlvp

**tlvp** is a CLI [TLV](https://en.wikipedia.org/wiki/Type-length-value) parser specially designed to handle EMV data. This may
be used by payment system researchers and practitioners to read TLV data (both
binary and hex-encoded) in a formatted, organised way.

## Usage

```
$ tlvp --help
tlvp is a TLV parser for EMV data

Usage:

	tlvp [--describe] [--force-color|--no-color] [--help] [input]

	--describe:
		Outputs extra descriptions of tags

	--force-color:
		Forces color formatting

	--help:
		Displays this message

	--no-color:
		Disables color formatting

	--pdol:
		Reads data as a PDOL.

	[input]:
		Path to a file to be parsed. Standard input can also be used by piping
		data into tlvp.

Bug reports:
	Please file an issue: https://github.com/victorgama/tlvp/issues/new

License:
	MIT. Copyright (C) 2020 - Victor Gama (hey@vito.io)


$ tlvp data.bin
[6F] File Control Information (FCI) Template
  │   Size: 64 bytes
  ├── [84] Dedicated File (DF) Name
  │   Size: 07 bytes
  ...


$ tlvp data.txt
[6F] File Control Information (FCI) Template
  │   Size: 64 bytes
  ├── [84] Dedicated File (DF) Name
  │   Size: 07 bytes
  ...


$ curl https://some.tlv.data.source | tlvp
[6F] File Control Information (FCI) Template
  │   Size: 64 bytes
  ├── [84] Dedicated File (DF) Name
  │   Size: 07 bytes
  ...


$ tlvp --describe data.txt
[6F] File Control Information (FCI) Template
  │   Identifies the FCI template according to ISO/IEC 7816-4
  │   Size: 64 bytes
  │
  ├── [84] Dedicated File (DF) Name
  │   Identifies the name of the DF as described in ISO/IEC 7816-4
  │   Size: 07 bytes
  ...
```

## Acknowledgments
Data included in [tags.json](tags.json) was kindly provided by 
[Dr. Steven J. Murdoch](https://murdoch.is/). All rights reserved. <br />
Such data is available at https://emvlab.org/emvtags/all/

## License

```
The MIT License (MIT)

Copyright (c) 2014-2015 Victor Gama

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

Data included in tags.json was kindly provided by Dr Steven J. Murdoch
(https://murdoch.is/). All rights reserved.
Data is available at https://emvlab.org/emvtags/all/
```
