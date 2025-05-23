import { useState, useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { useBreadcrumbs } from '@/contexts/BreadcrumbContext';
import { studyApi, type StudyActivity } from '@/lib/api';
import { Button } from '@/components/ui/button';
import { ArrowLeft, Check, X, HelpCircle, Clock, BarChart2, CheckCircle2, BookOpen } from 'lucide-react';
import { Skeleton } from '@/components/ui/skeleton';
import { Progress } from '@/components/ui/progress';

type Question = {
  id: number;
  question: string;
  options: string[];
  correctAnswer: number;
  explanation: string;
};

// Mock questions - in a real app, these would come from your backend
const MOCK_QUESTIONS: Question[] = [
  {
    id: 1,
    question: 'What is the past tense of "go"?',
    options: ['goed', 'gone', 'went', 'goes'],
    correctAnswer: 2,
    explanation: 'The past tense of "go" is "went". Example: "I went to the store yesterday."'
  },
  {
    id: 2,
    question: 'Which word is a synonym for "happy"?',
    options: ['Sad', 'Angry', 'Joyful', 'Tired'],
    correctAnswer: 2,
    explanation: '"Joyful" is a synonym for "happy" as both express a feeling of pleasure and contentment.'
  },
  {
    id: 3,
    question: 'Choose the correct sentence:',
    options: [
      'She don\'t like apples',
      'She doesn\'t likes apples',
      'She doesn\'t like apples',
      'She do not like apples'
    ],
    correctAnswer: 2,
    explanation: 'The correct form is "She doesn\'t like apples" because with third person singular (she/he/it), we use "does" + base form of the verb.'
  }
];

export default function StudyActivityLaunchPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { setItems } = useBreadcrumbs();
  
  const [activity, setActivity] = useState<StudyActivity | null>(null);
  const [currentQuestionIndex, setCurrentQuestionIndex] = useState(0);
  const [selectedOption, setSelectedOption] = useState<number | null>(null);
  const [showResult, setShowResult] = useState(false);
  const [isCorrect, setIsCorrect] = useState(false);
  const [score, setScore] = useState(0);
  const [timeRemaining, setTimeRemaining] = useState(300); // 5 minutes in seconds
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [completed, setCompleted] = useState(false);

  const currentQuestion = MOCK_QUESTIONS[currentQuestionIndex];
  const progress = ((currentQuestionIndex + 1) / MOCK_QUESTIONS.length) * 100;

  useEffect(() => {
    const fetchActivity = async () => {
      if (!id) return;
      
      try {
        setLoading(true);
        const data = await studyApi.getActivity(parseInt(id));
        setActivity(data);
        
        // Update breadcrumb
        setItems([
          { label: 'Dashboard', path: '/' },
          { label: 'Study Activities', path: '/study-activities' },
          { label: data.name, path: `/study-activities/${id}` },
          { label: 'Practice' }
        ]);
      } catch (err) {
        console.error('Error fetching activity:', err);
        setError('Failed to load activity. Please try again later.');
      } finally {
        setLoading(false);
      }
    };

    fetchActivity();

    // Set up timer
    const timer = setInterval(() => {
      setTimeRemaining(prev => {
        if (prev <= 1) {
          clearInterval(timer);
          handleComplete();
          return 0;
        }
        return prev - 1;
      });
    }, 1000);

    return () => clearInterval(timer);
  }, [id, setItems]);

  const handleOptionSelect = (optionIndex: number) => {
    if (showResult) return; // Prevent changing answer after submission
    setSelectedOption(optionIndex);
  };

  const handleSubmit = () => {
    if (selectedOption === null) return;
    
    const correct = selectedOption === currentQuestion.correctAnswer;
    setIsCorrect(correct);
    if (correct) {
      setScore(prev => prev + 1);
    }
    setShowResult(true);
  };

  const handleNext = () => {
    if (currentQuestionIndex < MOCK_QUESTIONS.length - 1) {
      setCurrentQuestionIndex(prev => prev + 1);
      setSelectedOption(null);
      setShowResult(false);
    } else {
      handleComplete();
    }
  };

  const handleComplete = () => {
    setCompleted(true);
  };

  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs < 10 ? '0' : ''}${secs}`;
  };

  if (loading) {
    return (
      <div className="space-y-6 max-w-3xl mx-auto">
        <Button variant="outline" size="sm" onClick={() => navigate(-1)} className="mb-6">
          <ArrowLeft className="h-4 w-4 mr-2" /> Back to Activity
        </Button>
        
        <div className="space-y-8">
          <div className="space-y-2">
            <Skeleton className="h-8 w-3/4" />
            <Skeleton className="h-4 w-1/2" />
          </div>
          
          <div className="space-y-4">
            <Skeleton className="h-6 w-1/4" />
            <div className="space-y-2">
              {[1, 2, 3, 4].map((i) => (
                <Skeleton key={i} className="h-12 w-full rounded-lg" />
              ))}
            </div>
          </div>
          
          <div className="flex justify-between items-center pt-4">
            <Skeleton className="h-10 w-24" />
            <Skeleton className="h-10 w-24" />
          </div>
        </div>
      </div>
    );
  }

  if (error || !activity) {
    return (
      <div className="space-y-4 max-w-3xl mx-auto">
        <Button variant="outline" size="sm" onClick={() => navigate(-1)}>
          <ArrowLeft className="h-4 w-4 mr-2" /> Back to Activity
        </Button>
        <div className="rounded-lg border border-destructive bg-destructive/10 p-4 text-destructive">
          <p>{error || 'Activity not found'}</p>
          <Button 
            variant="outline" 
            size="sm" 
            className="mt-2"
            onClick={() => window.location.reload()}
          >
            Retry
          </Button>
        </div>
      </div>
    );
  }

  if (completed) {
    const percentage = Math.round((score / MOCK_QUESTIONS.length) * 100);
    
    return (
      <div className="max-w-2xl mx-auto text-center py-12 px-4">
        <div className="bg-card border rounded-2xl p-8 shadow-sm">
          <div className="w-20 h-20 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-6">
            <CheckCircle2 className="h-10 w-10 text-green-600" />
          </div>
          <h1 className="text-3xl font-bold mb-2">Practice Complete!</h1>
          <p className="text-muted-foreground mb-8">You've completed the {activity.name} activity.</p>
          
          <div className="bg-muted/50 rounded-xl p-6 mb-8">
            <div className="grid grid-cols-3 gap-4 mb-6">
              <div>
                <p className="text-sm text-muted-foreground">Score</p>
                <p className="text-3xl font-bold">{score}/{MOCK_QUESTIONS.length}</p>
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Accuracy</p>
                <p className="text-3xl font-bold">{percentage}%</p>
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Time</p>
                <p className="text-3xl font-bold">{formatTime(300 - timeRemaining)}</p>
              </div>
            </div>
            
            <div className="space-y-2">
              <div className="flex justify-between text-sm">
                <span>Correct Answers</span>
                <span className="font-medium">{score} of {MOCK_QUESTIONS.length}</span>
              </div>
              <Progress value={(score / MOCK_QUESTIONS.length) * 100} className="h-2" />
            </div>
          </div>
          
          <div className="flex flex-col sm:flex-row gap-3 justify-center">
            <Button variant="outline" asChild>
              <Link to={`/study-activities/${id}`}>
                <BookOpen className="h-4 w-4 mr-2" /> Review Activity
              </Link>
            </Button>
            <Button asChild>
              <Link to="/study-activities">
                Back to Activities
              </Link>
            </Button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-3xl mx-auto">
      <div className="mb-6">
        <Button variant="outline" size="sm" onClick={() => navigate(-1)}>
          <ArrowLeft className="h-4 w-4 mr-2" /> Back to Activity
        </Button>
      </div>
      
      <div className="space-y-6">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold">{activity.name}</h1>
            <p className="text-muted-foreground">Question {currentQuestionIndex + 1} of {MOCK_QUESTIONS.length}</p>
          </div>
          
          <div className="flex items-center gap-4">
            <div className="flex items-center text-sm text-muted-foreground">
              <Clock className="h-4 w-4 mr-1.5" />
              {formatTime(timeRemaining)}
            </div>
            <div className="flex items-center text-sm text-muted-foreground">
              <BarChart2 className="h-4 w-4 mr-1.5" />
              {score} / {MOCK_QUESTIONS.length}
            </div>
          </div>
        </div>
        
        <Progress value={progress} className="h-2" />
        
        <div className="bg-card border rounded-xl p-6 shadow-sm">
          <div className="space-y-6">
            <div className="space-y-2">
              <div className="flex items-center text-sm text-muted-foreground mb-2">
                <HelpCircle className="h-4 w-4 mr-2" />
                Question {currentQuestionIndex + 1}
              </div>
              <h2 className="text-xl font-semibold">{currentQuestion.question}</h2>
            </div>
            
            <div className="space-y-3">
              {currentQuestion.options.map((option, index) => {
                let optionClass = "w-full text-left p-4 rounded-lg border hover:bg-accent transition-colors ";
                
                if (showResult) {
                  if (index === currentQuestion.correctAnswer) {
                    optionClass += "bg-green-50 border-green-200 text-green-900";
                  } else if (index === selectedOption && !isCorrect) {
                    optionClass += "bg-red-50 border-red-200 text-red-900";
                  } else {
                    optionClass += "bg-muted/30 border-muted";
                  }
                } else {
                  optionClass += selectedOption === index 
                    ? "border-primary bg-accent" 
                    : "border-muted hover:border-muted-foreground/20";
                }
                
                return (
                  <button
                    key={index}
                    className={optionClass}
                    onClick={() => handleOptionSelect(index)}
                    disabled={showResult}
                  >
                    <div className="flex items-center">
                      <div className={`flex-shrink-0 h-5 w-5 rounded-full border flex items-center justify-center mr-3 ${
                        showResult 
                          ? index === currentQuestion.correctAnswer 
                            ? 'bg-green-100 border-green-300 text-green-600' 
                            : index === selectedOption && !isCorrect 
                              ? 'bg-red-100 border-red-300 text-red-600' 
                              : 'bg-muted border-muted-foreground/20'
                          : selectedOption === index 
                            ? 'bg-primary text-primary-foreground border-primary' 
                            : 'bg-background border-muted-foreground/30'
                      }`}>
                        {showResult && index === currentQuestion.correctAnswer ? (
                          <Check className="h-3 w-3" />
                        ) : showResult && index === selectedOption && !isCorrect ? (
                          <X className="h-3 w-3" />
                        ) : null}
                      </div>
                      <span className="text-left">{option}</span>
                    </div>
                  </button>
                );
              })}
            </div>
            
            {showResult && (
              <div className={`p-4 rounded-lg ${
                isCorrect ? 'bg-green-50 border border-green-200' : 'bg-red-50 border border-red-200'
              }`}>
                <h3 className={`font-medium mb-1 ${
                  isCorrect ? 'text-green-800' : 'text-red-800'
                }`}>
                  {isCorrect ? 'Correct!' : 'Incorrect'}
                </h3>
                <p className="text-sm text-muted-foreground">
                  {currentQuestion.explanation}
                </p>
              </div>
            )}
            
            <div className="flex justify-between pt-2">
              <Button 
                variant="outline" 
                onClick={() => navigate(-1)}
                disabled={showResult && !isCorrect}
              >
                Exit
              </Button>
              
              {!showResult ? (
                <Button 
                  onClick={handleSubmit}
                  disabled={selectedOption === null}
                >
                  Check Answer
                </Button>
              ) : (
                <Button onClick={handleNext}>
                  {currentQuestionIndex < MOCK_QUESTIONS.length - 1 ? 'Next Question' : 'Finish'}
                </Button>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}