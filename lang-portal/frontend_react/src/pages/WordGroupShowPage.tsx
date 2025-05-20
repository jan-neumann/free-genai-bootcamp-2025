import { useParams } from 'react-router-dom';

export default function WordGroupShowPage() {
  const { id } = useParams();
  return (
    <div>
      <h1>Word Group Show Page</h1>
      <p>Group ID: {id}</p>
    </div>
  );
} 