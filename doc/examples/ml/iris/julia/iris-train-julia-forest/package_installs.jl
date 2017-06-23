metadata_packages = [
    "DataFrames",
    "DecisionTree",
    "HDF5",
    "JLD"]


Pkg.init()
Pkg.update()

for package=metadata_packages
    Pkg.add(package)
end

Pkg.resolve()
