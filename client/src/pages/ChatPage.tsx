import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { api, QuestionResponse } from '../lib/api';

export default function ChatPage() {
  const { sessionId } = useParams<{ sessionId: string }>();
  const navigate = useNavigate();

  const [question, setQuestion] = useState<QuestionResponse | null>(null);
  const [answer, setAnswer] = useState('');
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchNextQuestion = async () => {
    if (!sessionId) return;

    setLoading(true);
    setError(null);

    try {
      const nextQuestion = await api.getNextQuestion(sessionId);

      if (nextQuestion.done) {
        navigate(`/complete/${sessionId}`);
        return;
      }

      setQuestion(nextQuestion);
      setAnswer('');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch question');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchNextQuestion();
  }, [sessionId]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!sessionId || !question || !answer.trim()) return;

    setSubmitting(true);
    setError(null);

    try {
      await api.submitAnswer(sessionId, question.field, answer.trim());
      await fetchNextQuestion();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to submit answer');
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="text-white text-xl">Loading...</div>
      </div>
    );
  }

  if (!question) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="text-red-400 text-xl">No questions available</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background flex items-center justify-center p-6">
      <div className="max-w-2xl w-full">
        {/* Progress indicator */}
        <div className="mb-8">
          <div className="flex justify-between items-center mb-2">
            <span className="text-gray-400 text-sm">
              Question {question.progress + 1} of {question.total}
            </span>
            <span className="text-primary font-semibold text-sm">
              {Math.round(((question.progress) / question.total) * 100)}% Complete
            </span>
          </div>
          <div className="w-full bg-gray-800 rounded-full h-2">
            <div
              className="bg-primary h-2 rounded-full transition-all duration-300"
              style={{ width: `${((question.progress) / question.total) * 100}%` }}
            />
          </div>
        </div>

        {/* Chat interface */}
        <div className="bg-gray-900 rounded-2xl p-8 shadow-2xl">
          <div className="mb-6">
            <div className="inline-block bg-primary bg-opacity-10 text-primary px-3 py-1 rounded-full text-xs font-semibold mb-3">
              {question.isAIPhrased ? 'AI-Generated' : 'Auto-Generated'}
            </div>
            <h2 className="text-2xl font-semibold text-white mb-2">
              {question.question}
            </h2>
            <p className="text-gray-400 text-sm">Field: {question.field}</p>
          </div>

          <form onSubmit={handleSubmit}>
            <div className="mb-6">
              <textarea
                value={answer}
                onChange={(e) => setAnswer(e.target.value)}
                placeholder="Type your answer here..."
                rows={4}
                className="w-full bg-gray-800 text-white border border-gray-700 rounded-xl px-4 py-3 focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent resize-none"
                disabled={submitting}
                autoFocus
              />
            </div>

            {error && (
              <div className="mb-4 bg-red-900 bg-opacity-20 border border-red-700 text-red-400 px-4 py-3 rounded-lg text-sm">
                {error}
              </div>
            )}

            <button
              type="submit"
              disabled={!answer.trim() || submitting}
              className={`w-full py-4 px-6 rounded-xl font-semibold text-white transition-all ${
                !answer.trim() || submitting
                  ? 'bg-gray-700 cursor-not-allowed'
                  : 'bg-primary hover:bg-opacity-90'
              }`}
            >
              {submitting ? 'Submitting...' : 'Next →'}
            </button>
          </form>
        </div>

        {/* Helper text */}
        <p className="mt-4 text-center text-gray-500 text-sm">
          Press Enter while holding Shift to add a new line
        </p>
      </div>
    </div>
  );
}
