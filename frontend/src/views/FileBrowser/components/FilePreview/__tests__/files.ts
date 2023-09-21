const files = {
  c: `
#include <stdio.h>

int main() {
    printf("Hello, World!");
    return 0;
}
  `,
  cpp: `
#include <iostream>

int main() {
    std::cout << "Hello World!";
    return 0;
}
  `,
  css: `
.codePreview {
  border: solid 1px var(--grey);
  margin: 2rem auto;
  max-width: 1000px;
}
  `,
  csv: `hello,world\n1,2\n`,
  docker: `
ARG DOCKER_TAG=\${DOCKER_TAG:-local}
ENV REACT_APP_RELEASE_VERSION=\${DOCKER_TAG:-local}

WORKDIR /usr/src/app

COPY --from=0 /usr /usr
# Start the server
CMD ["npm", "start", "--prefix", "./backend"]
EXPOSE 4000
`,
  go: `
package main

import "fmt"

func main() {
  fmt.Println("hello world")
}
`,
  java: `
class HelloWorld {
  public static void main(String[] args) {
      System.out.println("Hello, World!");
  }
}
`,
  javascript: `
import React from 'react';

type HeaderProps = {
  text: string;
};

const Header = ({text}: HeaderProps) => {
  return <h1>{text}</h1>;
};

export default Header;
  `,
  json: `
{
  "hello": "world"
}
  `,
  julia: `
s1 = "Hello World!"
println(s1)
  `,
  html: `
<!DOCTYPE html>
  <html>
    <body>
      hello world
    </body>
  </html>
  `,
  markdown: `
# H1
## H2
### H3
---
**labore et dolore magna aliqua**
  `,
  protobuf: `
syntax = "proto3";

package pfs_v2;
option go_package = "github.com/pachyderm/pachyderm/v2/src/pfs";

import "protoextensions/validate.proto";

////  PFS Data structures (stored in etcd)

  message Repo {
    string name = 1;
    string type = 2;
    Project project = 3;
  }

`,
  php: `
<!DOCTYPE html>
<html>
  <head>
    <title>PHP Test</title>
  </head>
  <body>
    <?php echo '<p>Hello World</p>'; ?>
  </body>
</html>
`,
  python: `
x = 1
if x == 1:
  print(x)
  `,
  r: `
# Print Hello World
message <- "Hello, World!"
print(message)
`,
  ruby: `
=begin
Weird multi-line
comments in here
=end

puts 'Hello, world!'

`,
  rs: `
fn main() {
  // Print text to the console.
  println!("Hello World!");
}
  `,
  shell: `
#!/bin/bash

set -euo pipefail
IFS=$'\n\t'

COMPONENT_LIBARY_CHANGED=$(git diff --name-status HEAD~1...HEAD frontend/components)

if [ "$COMPONENT_LIBARY_CHANGED" ]; then
    cd frontend
    npm ci
    npm run storybook:publish
else
    echo 'Skipped storybook upload'
fi
`,
  sql: `
CREATE TABLE helloworld (phrase TEXT);
INSERT INTO helloworld VALUES ("Hello, World!");
INSERT INTO helloworld VALUES ("Goodbye, World!");
SELECT COUNT(*) FROM helloworld;
  `,
  svg: `
<svg height="100" width="100" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
  <circle cx="50" cy="50" r="40" fill="black" />
</svg>
  `,
  tsv: `hello\tworld\n1\t2\n`,
  xml: `
<note>
  <to>Aerith</to>
  <from>Cloud</from>
  <body volume="loud">Look behind you!</body>
</note>
  `,
  yaml: `
hello: world
  `,
};

export default files;
