# OpenCV Example in Go

This is the OpenCV example that uses the Pachyderm Go API.
To run this example, you must have the [Go](https://golang.org/)
programming language installed and configured on your computer.

To run the OpenCV example, execute `opencv-example.go` from the
root of the Pachyderm repository.

**Example:**

```
go run examples/opencv/goclient-example/opencv-example.go
```

**System Response:**

```
[file:<commit:<repo:<name:"images" > id:"02f3497072ef48939b9ad936bbc28df2" > path:"/AT-AT.png" > file_type:FILE size_bytes:80588 committed:<seconds:1585938116 nanos:958561905 > hash:"\331\377s\\\323Fv1\251\353\312\347M&\320\207\321\254eT\347\320\025o\242k\371\311\224\352fh"  file:<commit:<repo:<name:"images" > id:"02f3497072ef48939b9ad936bbc28df2" > path:"/kitten.png" > file_type:FILE size_bytes:104836 committed:<seconds:1585938116 nanos:958561905 > hash:"\323j\321M\007\260|\245v%|ZS\274\317[\372\3040\275\356#\360\313\213\345\346=\274\247\264K"  file:<commit:<repo:<name:"images" > id:"02f3497072ef48939b9ad936bbc28df2" > path:"/liberty.png" > file_type:FILE size_bytes:58644 committed:<seconds:1585938116 nanos:958561905 > hash:"\341H\347\235\370\301WoUFw3\317K\232>\253\3432iU\253\267\245\213%J,I\365X\023" ]
[pipeline:<name:"montage" > version:1 transform:<image:"v4tech/imagemagick" cmd:"sh" stdin:"montage -shadow -background SkyBlue -geometry 300x300+2+2 $(find /pfs -type f | sort) /pfs/out/montage.png" > parallelism_spec:<constant:1 > created_at:<seconds:1585938117 nanos:158996551 > output_branch:"master" input:<cross:<pfs:<name:"edges" repo:"edges" branch:"master" glob:"/" > > cross:<pfs:<name:"images" repo:"images" branch:"master" glob:"/" > > > cache_size:"64M" salt:"dbb4450deb1a4e48b9b351716c106c17" max_queue_size:1 spec_commit:<repo:<name:"__spec__" > id:"65d065443358471599b5e2c818b73e9a" > datum_tries:3  pipeline:<name:"edges" > version:1 transform:<image:"pachyderm/opencv" cmd:"python3" cmd:"/edges.py" > parallelism_spec:<constant:1 > created_at:<seconds:1585938116 nanos:968483073 > output_branch:"master" input:<pfs:<name:"images" repo:"images" branch:"master" glob:"/*" > > cache_size:"64M" salt:"6388bc3ff5244b4a8efff24fff49018e" max_queue_size:1 spec_commit:<repo:<name:"__spec__" > id:"d610d27e68b7405198c64e9fbbd88e78" > datum_tries:3 ]
```

The example creates the following:

- Pipelines:

  - `edges`
  - `montage`

- Repositories:

  - `images`
  - `edges`
  - `montage`
