library(caret)
library(tools)

# load model
load("/pfs/model/model.rda")

# loop over input files performing inference
files <- list.files(path = "/pfs/attributes", 
                    pattern = "*.csv", 
                    full.names = T, 
                    recursive = FALSE)
					
cols <- c("Sepal.Length",
          "Sepal.Width",
          "Petal.Length",
          "Petal.Width")

species <- c("setosa",
             "versicolor",
             "virginica")
		  
for (file in files){
	# load the data
	dataset <- read.csv(file, header=FALSE)
        colnames(dataset) <- cols
        
        # perform the inference
        predictions <- predict(fit.model, dataset)
        
	# output the results
        write(species[predictions], 
              file = paste("/pfs/out/", file_path_sans_ext(basename(file)), sep = "")) 
}

