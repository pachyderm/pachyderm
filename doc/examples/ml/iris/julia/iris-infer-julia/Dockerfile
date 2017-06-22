FROM julia

# Install packages.
ADD package_installs.jl /tmp/package_installs.jl
RUN apt-get update && \
    apt-get install -y build-essential hdf5-tools && \
    julia /tmp/package_installs.jl && \
    rm -rf /var/lib/apt/lists/*

# Add our program.
ADD infer.jl /infer.jl
