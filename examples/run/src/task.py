import os

# Get secret credentials
database_username = os.getenv("DATABASE_USERNAME", "username")
database_password = os.getenv("DATABASE_PASSWORD", "password")
database_host = os.getenv("DATABASE_HOST", "localhost")
database_port = os.getenv("DATABASE_PORT", 3306)

# Pipeline input output folders that are mapped to repositories or test folders
in_folder = os.getenv("PIPELINE_INPUT")
out_folder = os.getenv("PIPELINE_OUTPUT")
pipeline_home_folder = os.getenv("PIPELINE_HOME")

print(
    "In theory, the task should connect to {}:{}@{}:{}".format(
        database_username, database_password, database_host, database_port
    )
)
print("Furthermore, data should be moved from {} to {}".format(in_folder, out_folder))
