import React, { useState } from 'react';
import { useOutletContext } from 'react-router-dom';
import type { LayoutContextType } from '../components/DashboardLayout';

interface StorageFile {
  id: string;
  name: string;
  size: number;
  type: string;
  uploadedAt: string;
}

const INITIAL_FILES: StorageFile[] = [
  {
    id: '1',
    name: 'user_avatars_zip_pack.zip',
    size: 4529182, // ~4.3MB
    type: 'application/zip',
    uploadedAt: new Date(Date.now() - 3600000 * 24 * 2).toISOString(), // 2 days ago
  },
  {
    id: '2',
    name: 'production_banner_logo.webp',
    size: 245100, // ~240KB
    type: 'image/webp',
    uploadedAt: new Date(Date.now() - 3600000 * 4).toISOString(), // 4 hours ago
  },
  {
    id: '3',
    name: 'cors_gateway_rules.json',
    size: 1450, // ~1.4KB
    type: 'application/json',
    uploadedAt: new Date(Date.now() - 600000).toISOString(), // 10 mins ago
  },
];

export const Storage: React.FC = () => {
  const { showToast } = useOutletContext<LayoutContextType>();
  const [files, setFiles] = useState<StorageFile[]>(INITIAL_FILES);
  const [isDragging, setIsDragging] = useState(false);

  // Format Bytes utility
  const formatBytes = (bytes: number): string => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  // Add new file handler
  const handleAddFile = (file: File) => {
    const newFile: StorageFile = {
      id: Math.random().toString(36).substring(2, 9),
      name: file.name,
      size: file.size,
      type: file.type || 'application/octet-stream',
      uploadedAt: new Date().toISOString(),
    };
    setFiles((prev) => [newFile, ...prev]);
    showToast(`Successfully uploaded ${file.name} to MinIO S3 bucket.`, 'success');
  };

  // Dropzone drag over
  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(true);
  };

  // Dropzone drag leave
  const handleDragLeave = () => {
    setIsDragging(false);
  };

  // Dropzone drop
  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
    const file = e.dataTransfer.files[0];
    if (file) {
      handleAddFile(file);
    }
  };

  // Manual File input change
  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      handleAddFile(file);
    }
  };

  // Delete file
  const handleDeleteFile = (id: string, name: string) => {
    setFiles((prev) => prev.filter((f) => f.id !== id));
    showToast(`Removed ${name} from storage bucket.`, 'success');
  };

  // Calculate stats
  const totalSize = files.reduce((acc, f) => acc + f.size, 0);

  return (
    <div className="space-y-lg animate-fadeIn w-full">
      {/* Page Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-md mb-md">
        <div>
          <h2 className="font-headline-lg text-headline-lg font-bold text-[#191c1d] mb-xs">
            Object Storage
          </h2>
          <p className="font-body-md text-body-md text-[#464554]">
            Store, manage, and process static assets. Includes automatic Lanczos3 image resizing.
          </p>
        </div>

        {/* Upload file triggers */}
        <label className="bg-[#4648d4] hover:bg-[#6063ee] text-white py-sm px-lg rounded-lg font-label-md text-label-md flex items-center justify-center space-x-sm shrink-0 shadow-sm cursor-pointer transition-colors focus-within:ring-2 focus-within:ring-[#4648d4]/50">
          <span className="material-symbols-outlined">upload_file</span>
          <span>Upload File</span>
          <input
            type="file"
            onChange={handleFileChange}
            className="sr-only"
            aria-label="Upload file manually"
          />
        </label>
      </div>

      {/* Grid: Dropzone + Stats */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-lg">
        
        {/* Left Column: Dropzone & Upload card (Aligned and Centered) */}
        <div className="lg:col-span-2 space-y-md">
          <div
            onDragOver={handleDragOver}
            onDragLeave={handleDragLeave}
            onDrop={handleDrop}
            className={`
              border-2 border-dashed rounded-xl p-xl flex flex-col items-center justify-center min-h-[300px] text-center transition-all duration-200 cursor-pointer
              ${isDragging
                ? 'border-[#4648d4] bg-[#e1e0ff]/20 text-[#07006c]'
                : 'border-[#c7c4d7] bg-white hover:bg-[#f8f9fa] text-[#464554]'
              }
            `}
          >
            <span className={`material-symbols-outlined text-5xl mb-md transition-transform ${isDragging ? 'scale-110 text-[#4648d4]' : 'text-[#4648d4]/40'}`}>
              cloud_upload
            </span>
            <h3 className="font-headline-md text-lg text-[#191c1d] font-bold mb-xs">
              Drag and drop files here
            </h3>

            
            <label className="inline-block bg-[#e2dfff] hover:bg-[#c3c0ff] text-[#0f0069] px-xl py-sm rounded-lg text-sm font-semibold transition-colors cursor-pointer focus-within:ring-2 focus-within:ring-[#4648d4]/30">
              <span>Browse Files</span>
              <input
                type="file"
                onChange={handleFileChange}
                className="hidden"
                aria-label="Browse files to upload"
              />
            </label>
          </div>
        </div>

        {/* Right Column: Bucket Statistics Panel */}
        <div className="bg-white rounded-xl shadow-sm p-lg border border-[#c7c4d7]/30 flex flex-col justify-between space-y-md">
          <div>
            <h3 className="font-headline-md text-headline-md font-bold text-[#191c1d] mb-sm">
              Bucket Statistics
            </h3>
            <p className="font-body-sm text-xs text-[#464554] mb-md leading-normal">
              Active configuration parameters for bucket: <strong>strata-public-bucket</strong>.
            </p>

            <div className="space-y-sm text-sm border-t border-[#c7c4d7]/20 pt-md">
              <div className="flex justify-between border-b border-[#c7c4d7]/10 pb-xs">
                <span className="text-[#464554]">Total Files</span>
                <span className="font-semibold text-[#191c1d]">{files.length} items</span>
              </div>
              <div className="flex justify-between border-b border-[#c7c4d7]/10 pb-xs">
                <span className="text-[#464554]">Bucket Size</span>
                <span className="font-semibold text-[#191c1d]">{formatBytes(totalSize)}</span>
              </div>
              <div className="flex justify-between border-b border-[#c7c4d7]/10 pb-xs">
                <span className="text-[#464554]">Storage Engine</span>
                <span className="font-semibold font-code-sm">MinIO S3 (Go-wrapper)</span>
              </div>
              <div className="flex justify-between pb-xs">
                <span className="text-[#464554]">Image Resizing</span>
                <span className="font-semibold text-[#0a5c0a] flex items-center gap-xs">
                  <span className="w-1.5 h-1.5 rounded-full bg-[#0a5c0a]" />
                  <span>Lanczos3 Enabled</span>
                </span>
              </div>
            </div>
          </div>

          <div className="bg-[#e1e0ff]/30 text-[#07006c] border border-[#c0c1ff]/60 p-sm rounded-lg text-xs leading-normal">
            <strong>Note:</strong> All static images uploaded are served via optimized caching policies automatically configured at the API Gateway.
          </div>
        </div>
      </div>

      {/* Uploaded Files Table List */}
      <div className="space-y-md">
        <h3 className="font-headline-md text-headline-md font-bold text-[#191c1d]">
          Assets Inventory
        </h3>

        {files.length === 0 ? (
          /* Centers and properly renders empty table state */
          <div className="bg-white rounded-xl shadow-sm border border-[#c7c4d7]/30 p-xl flex flex-col items-center justify-center text-center min-h-[200px]">
            <span className="material-symbols-outlined text-4xl text-[#464554]/30 mb-sm">folder_open</span>
            <p className="font-body-md text-body-md text-[#464554]">
              No files in the storage inventory yet.
            </p>
          </div>
        ) : (
          <div className="bg-white rounded-xl shadow-sm border border-[#c7c4d7]/30 overflow-hidden">
            <div className="overflow-x-auto w-full">
              <table className="w-full border-collapse text-left">
                <thead>
                  <tr className="bg-[#f3f4f5] border-b border-[#c7c4d7]/30 text-[#464554] uppercase font-label-md text-label-md tracking-wider">
                    <th className="px-lg py-md">File Name</th>
                    <th className="px-lg py-md w-36">Size</th>
                    <th className="px-lg py-md w-44">Content Type</th>
                    <th className="px-lg py-md w-44">Uploaded At</th>
                    <th className="px-lg py-md w-28 text-center">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-[#c7c4d7]/20 font-body-md text-body-md text-[#191c1d]">
                  {files.map((file) => (
                    <tr key={file.id} className="hover:bg-[#f8f9fa] transition-colors">
                      <td className="px-lg py-md font-semibold truncate max-w-sm flex items-center gap-sm">
                        <span className="material-symbols-outlined text-[#4648d4]">
                          {file.type.startsWith('image/')
                            ? 'image'
                            : file.type.includes('zip')
                            ? 'zip_box'
                            : 'description'
                          }
                        </span>
                        <span className="truncate">{file.name}</span>
                      </td>
                      <td className="px-lg py-md text-[#464554] font-code-sm text-code-sm">
                        {formatBytes(file.size)}
                      </td>
                      <td className="px-lg py-md text-[#464554] font-code-sm text-code-sm truncate max-w-[150px]">
                        {file.type}
                      </td>
                      <td className="px-lg py-md text-[#464554] text-sm whitespace-nowrap">
                        {new Date(file.uploadedAt).toLocaleString([], { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })}
                      </td>
                      <td className="px-lg py-md text-center">
                        <button
                          onClick={() => handleDeleteFile(file.id, file.name)}
                          className="text-[#ba1a1a] hover:bg-[#ffdad6]/40 p-xs rounded transition-colors cursor-pointer focus:outline-none focus:ring-2 focus:ring-[#ba1a1a]/30"
                          title="Delete File"
                        >
                          <span className="material-symbols-outlined text-lg">delete</span>
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
