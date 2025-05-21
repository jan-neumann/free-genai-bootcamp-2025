import { Button } from "@/components/ui/button";
import { ArrowRight } from "lucide-react";
import { Link } from "react-router-dom";

export default function StartStudyingButton() {
  return (
    <div className="flex justify-center">
      <Link to="/study-activities">
        <Button className="w-full">
          Start Studying
          <ArrowRight className="ml-2 h-4 w-4" />
        </Button>
      </Link>
    </div>
  );
}
