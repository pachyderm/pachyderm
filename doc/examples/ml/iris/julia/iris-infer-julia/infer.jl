using DataFrames
using DecisionTree
using JLD

# Load our model.
model = load(ARGS[1], "model")

# Walk over the directory with input attribute files.
attributes = readdir(ARGS[2])
for file in attributes
  p = joinpath(ARGS[2], file)
  if isdir(p)
    continue
  elseif isfile(p)
    df = readtable(p, header = false)
    open(joinpath(ARGS[3], file), "a") do x
      for r in eachrow(df)
        prediction = DecisionTree.predict(model, convert(Array, r))
        write(x, string(prediction[1], "\n"))
      end
    end
  end
end

