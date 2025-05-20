import { useParams } from 'react-router-dom';

export default function StudyActivityLaunchPage() {
  const { id } = useParams();
  return (
    <div>
      <h1>Study Activity Launch Page</h1>
      <p>Activity ID: {id}</p>
    </div>
  );
} 