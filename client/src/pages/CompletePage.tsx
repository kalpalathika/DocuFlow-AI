import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { api, type SessionStatusResponse } from '../lib/api';

export default function CompletePage() {
  const { sessionId } = useParams<{ sessionId: string }>();
  const navigate = useNavigate();

  const [session, setSession] = useState<SessionStatusResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [downloading, setDownloading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [editingField, setEditingField] = useState<string | null>(null);
  const [editValue, setEditValue] = useState<string>('');
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    const fetchSession = async () => {
      if (!sessionId) return;

      try {
        const sessionData = await api.getSession(sessionId);
        setSession(sessionData);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load session');
      } finally {
        setLoading(false);
      }
    };

    fetchSession();
  }, [sessionId]);

  const handleEdit = (field: string, currentValue: string) => {
    setEditingField(field);
    setEditValue(currentValue);
  };

  const handleCancelEdit = () => {
    setEditingField(null);
    setEditValue('');
  };

  const handleSaveEdit = async () => {
    if (!sessionId || !editingField) return;

    setSaving(true);
    setError(null);

    try {
      await api.submitAnswer(sessionId, editingField, editValue);

      // Update local state
      if (session) {
        setSession({
          ...session,
          answers: {
            ...session.answers,
            [editingField]: editValue,
          },
        });
      }

      setEditingField(null);
      setEditValue('');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save changes');
    } finally {
      setSaving(false);
    }
  };

  const handleDownload = async () => {
    if (!sessionId) return;

    setDownloading(true);
    setError(null);

    try {
      const blob = await api.downloadDocument(sessionId);

      // Create a download link
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `document-${sessionId}.docx`;
      document.body.appendChild(a);
      a.click();

      // Cleanup
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to download document');
    } finally {
      setDownloading(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="text-white text-xl">Loading...</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background py-8 px-6">
      <div className="max-w-3xl mx-auto">
        {/* Success Header */}
        <div className="bg-gray-900 rounded-2xl p-6 shadow-2xl text-center mb-6">
          <div className="mb-4">
            <div className="mx-auto w-16 h-16 bg-primary bg-opacity-20 rounded-full flex items-center justify-center">
              <svg
                className="w-8 h-8 text-primary"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M5 13l4 4L19 7"
                />
              </svg>
            </div>
          </div>

          <h1 className="text-3xl font-bold text-white mb-2">
            Document Ready!
          </h1>

          <p className="text-gray-400">
            Review your answers below and download your completed document
          </p>
        </div>

        {/* Preview Section */}
        <div className="bg-gray-900 rounded-2xl p-6 shadow-2xl mb-6">
          <h2 className="text-xl font-semibold text-white mb-4">
            Document Preview
          </h2>

          <div className="space-y-3">
            {session && Object.entries(session.answers).map(([field, answer]) => (
              <div key={field} className="bg-gray-800 rounded-lg p-4">
                <div className="flex justify-between items-start mb-2">
                  <div className="text-sm text-gray-400">
                    {field.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())}
                  </div>
                  {editingField !== field && (
                    <button
                      onClick={() => handleEdit(field, answer)}
                      className="text-primary hover:text-opacity-80 text-sm font-semibold"
                    >
                      Edit
                    </button>
                  )}
                </div>

                {editingField === field ? (
                  <div>
                    <textarea
                      value={editValue}
                      onChange={(e) => setEditValue(e.target.value)}
                      rows={3}
                      className="w-full bg-gray-900 text-white border border-gray-700 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent resize-none mb-3"
                      autoFocus
                    />
                    <div className="flex gap-2">
                      <button
                        onClick={handleSaveEdit}
                        disabled={saving}
                        className={`flex-1 py-2 px-4 rounded-lg font-semibold text-white transition-all ${
                          saving
                            ? 'bg-gray-700 cursor-not-allowed'
                            : 'bg-primary hover:bg-opacity-90'
                        }`}
                      >
                        {saving ? 'Saving...' : 'Save'}
                      </button>
                      <button
                        onClick={handleCancelEdit}
                        disabled={saving}
                        className="flex-1 py-2 px-4 rounded-lg font-semibold text-gray-300 bg-gray-700 hover:bg-gray-600 transition-all"
                      >
                        Cancel
                      </button>
                    </div>
                  </div>
                ) : (
                  <div className="text-white">
                    {answer || <span className="text-gray-500 italic">No answer provided</span>}
                  </div>
                )}
              </div>
            ))}
          </div>

          {session && Object.keys(session.answers).length === 0 && (
            <p className="text-gray-500 text-center py-8">No fields filled</p>
          )}
        </div>

        {error && (
          <div className="mb-6 bg-red-900 bg-opacity-20 border border-red-700 text-red-400 px-4 py-3 rounded-lg text-sm">
            {error}
          </div>
        )}

        {/* Action Buttons */}
        <div className="bg-gray-900 rounded-2xl p-6 shadow-2xl">
          <div className="space-y-3">
            <button
              onClick={handleDownload}
              disabled={downloading}
              className={`w-full py-4 px-6 rounded-xl font-semibold text-white transition-all ${
                downloading
                  ? 'bg-gray-700 cursor-not-allowed'
                  : 'bg-primary hover:bg-opacity-90'
              }`}
            >
              {downloading ? 'Generating Document...' : 'Download Document'}
            </button>

            <button
              onClick={() => navigate('/')}
              className="w-full py-4 px-6 rounded-xl font-semibold text-gray-300 bg-gray-800 hover:bg-gray-700 transition-all"
            >
              Start New Document
            </button>
          </div>

          <div className="mt-6 pt-6 border-t border-gray-800 text-center">
            <p className="text-gray-500 text-sm">
              Thank you for using DocuFlow AI
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
