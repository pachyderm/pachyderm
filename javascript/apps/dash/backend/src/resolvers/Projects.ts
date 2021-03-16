import client from '@dash-backend/grpc/client';
import {QueryResolvers} from '@graphqlTypes';

interface ProjectsResolver {
  Query: {
    projects: QueryResolvers['projects'];
  };
}

const projectsResolver: ProjectsResolver = {
  Query: {
    projects: async (
      _field,
      _args,
      {pachdAddress = '', authToken = '', log},
    ) => {
      const pachClient = client({pachdAddress, authToken, log});
      const projects = await pachClient.projects().listProject();
      return projects.projectInfoList.map((project) => {
        return {
          name: project.name,
          description: project.description,
          createdAt: project.createdat?.seconds || 0,
          status: project.status,
          id: project.id,
        };
      });
    },
  },
};

export default projectsResolver;
