import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { api } from '../lib/api';

export default function UploadPage() {
  const [file, setFile] = useState<File | null>(null);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const selectedFile = e.target.files?.[0];
    if (selectedFile) {
      if (!selectedFile.name.toLowerCase().endsWith('.docx')) {
        setError('Please upload a .docx file');
        setFile(null);
        return;
      }
      setFile(selectedFile);
      setError(null);
    }
  };

  const handleUpload = async () => {
    if (!file) {
      setError('Please select a file');
      return;
    }

    setUploading(true);
    setError(null);

    try {
      const response = await api.uploadDocument(file);
      navigate(`/chat/${response.sessionId}`);
    } catch (err) {
      if (err instanceof Error) {
        setError(err.message || 'Upload failed');
      } else {
        setError('Upload failed');
      }
    } finally {
      setUploading(false);
    }
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    const droppedFile = e.dataTransfer.files[0];
    if (droppedFile) {
      if (!droppedFile.name.toLowerCase().endsWith('.docx')) {
        setError('Please upload a .docx file');
        return;
      }
      setFile(droppedFile);
      setError(null);
    }
  };

  return (
    <div className="min-h-screen bg-background flex items-center justify-center p-6">
      <div className="max-w-xl w-full bg-gray-900 rounded-2xl p-8 shadow-2xl">
        <div className="mb-8">
          <h1 className="text-4xl font-bold text-white mb-2">DocuFlow AI</h1>
          <p className="text-gray-400">
            Upload your .docx template to get started
          </p>
        </div>

        <div
          onDragOver={handleDragOver}
          onDrop={handleDrop}
          onClick={() => document.getElementById('file-input')?.click()}
          className={`border-2 border-dashed rounded-xl p-12 text-center cursor-pointer transition-all ${
            file
              ? 'border-primary bg-primary bg-opacity-5'
              : 'border-gray-700 hover:border-gray-600 bg-gray-800'
          }`}
        >
          <input
            id="file-input"
            type="file"
            accept=".docx"
            onChange={handleFileChange}
            className="hidden"
          />

          <svg
            className="mx-auto mb-4 text-gray-500"
            width="48"
            height="48"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"
            />
          </svg>

          {file ? (
            <div>
              <p className="font-semibold text-white mb-1">{file.name}</p>
              <p className="text-sm text-gray-400">Click to change file</p>
            </div>
          ) : (
            <div>
              <p className="font-semibold text-white mb-1">
                Drop your .docx file here
              </p>
              <p className="text-sm text-gray-400">or click to browse</p>
            </div>
          )}
        </div>

        {error && (
          <div className="mt-4 bg-red-900 bg-opacity-20 border border-red-700 text-red-400 px-4 py-3 rounded-lg text-sm">
            {error}
          </div>
        )}

        <button
          onClick={handleUpload}
          disabled={!file || uploading}
          className={`mt-6 w-full py-4 px-6 rounded-xl font-semibold text-white transition-all ${
            !file || uploading
              ? 'bg-gray-700 cursor-not-allowed'
              : 'bg-primary hover:bg-opacity-90'
          }`}
        >
          {uploading ? 'Uploading...' : 'Start Filling Document'}
        </button>

        <div className="mt-8 bg-gray-800 rounded-xl p-5 text-sm text-gray-300">
          <p className="font-semibold text-white mb-3">How it works:</p>
          <ol className="space-y-2 pl-5 list-decimal">
            <li>Upload your .docx template with placeholders like {`{{client_name}}`}</li>
            <li>Answer questions in a conversational interface</li>
            <li>Download your completed document</li>
          </ol>
        </div>
      </div>
    </div>
  );
}
