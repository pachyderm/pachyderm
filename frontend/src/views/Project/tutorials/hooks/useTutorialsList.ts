import useLocalProjectSettings from '@dash-frontend/hooks/useLocalProjectSettings';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {Story} from '@pachyderm/components';

import useImageProcessingCleanup from '../ImageProcessing/hooks/useImageProcessingCleanup';
import ImageProcessingStories from '../ImageProcessing/stories';

const useTutorialsList = () => {
  const {projectId} = useUrlState();
  const imageProcessingCleanup = useImageProcessingCleanup();
  const [tutorialProgress] = useLocalProjectSettings({
    projectId,
    key: 'tutorial_progress',
  });

  const getTutorialProgress = (tutorialId: string, stories: Story[]) => {
    const currentStory =
      tutorialProgress && tutorialProgress[tutorialId]
        ? tutorialProgress[tutorialId].story
        : 0;
    const completedTask =
      tutorialProgress && tutorialProgress[tutorialId]
        ? tutorialProgress[tutorialId].task
        : -1;

    let initialTask = 0;
    let initialStory = 0;

    if (
      stories[currentStory]?.sections.filter((s) => !!s.Task).length ===
      completedTask + 1
    ) {
      if (currentStory + 1 <= stories.length - 1) {
        initialStory = currentStory + 1;
      } else {
        initialStory = currentStory;
        initialTask = completedTask;
      }
    } else {
      initialStory = currentStory;
      initialTask = completedTask + 1;
    }
    return {initialTask, initialStory};
  };

  return {
    'image-processing': {
      id: 'image-processing',
      content: 'Image processing at scale with pachyderm',
      cleanup: imageProcessingCleanup,
      stories: ImageProcessingStories,
      progress: getTutorialProgress('image-processing', ImageProcessingStories),
    },
  };
};

export default useTutorialsList;
