import useImageProcessingCleanup from '../ImageProcessing/hooks/useImageProcessingCleanup';
import ImageProcessingStories from '../ImageProcessing/stories';

const useTutorialsList = () => {
  const imageProcessingCleanup = useImageProcessingCleanup();

  return [
    {
      id: 'image-processing',
      content: 'Image processing at scale with pachyderm',
      cleanup: imageProcessingCleanup,
      stories: ImageProcessingStories.length,
    },
  ];
};

export default useTutorialsList;
