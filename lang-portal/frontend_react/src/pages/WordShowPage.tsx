import { useParams } from 'react-router-dom';

export default function WordShowPage() {
  const { id } = useParams();
  return (
    <div>
      <h1>Word Show Page</h1>
      <p>Word ID: {id}</p>
    </div>
  );
} 