package testing

var loads = []string{`
count: 3
operations:
  - count: 5
    operation:
      - putFile:
          files:
            count: 5
            file:
              - source: "random"
                prob: 1
        prob: 0.7
      - deleteFile:
          count: 5
          directoryProb: 0.2
        prob: 0.3
validator: {}
fileSources:
  - name: "random"
    random:
      directory:
        depth: 3
        run: 3
      size:
        - min: 1000
          max: 10000
          prob: 0.3
        - min: 10000
          max: 100000
          prob: 0.3
        - min: 1000000
          max: 10000000
          prob: 0.3
        - min: 10000000
          max: 100000000
          prob: 0.1
`, `
count: 3
operations:
  - count: 5
    operation:
      - putFile:
          files:
            count: 10000 
            file:
              - source: "random"
                prob: 1
        prob: 1
validator: {}
fileSources:
  - name: "random"
    random:
      size:
        - min: 100
          max: 1000
          prob: 1 
`, `
count: 3
operations:
  - count: 5
    operation:
      - putFile:
          files:
            count: 1
            file:
              - source: "random"
                prob: 1
        prob: 1
validator: {}
fileSources:
  - name: "random"
    random:
      size:
        - min: 10000000
          max: 100000000
          prob: 1 
`}
