"use client";

/**
 * LogoUpload Component
 *
 * Drag-and-drop logo upload with preview.
 * Validates file type (jpg, png) and size (max 2MB).
 */

import { useState, useCallback } from "react";
import Image from "next/image";
import { useDropzone } from "react-dropzone";
import { Upload, X, Loader2, Image as ImageIcon } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { useUploadLogoMutation } from "@/store/services/companyApi";

interface LogoUploadProps {
  currentLogoUrl?: string | null;
}

const MAX_FILE_SIZE = 2 * 1024 * 1024; // 2MB
const ACCEPTED_FILE_TYPES = {
  "image/jpeg": [".jpg", ".jpeg"],
  "image/png": [".png"],
};

export function LogoUpload({ currentLogoUrl }: LogoUploadProps) {
  const [preview, setPreview] = useState<string | null>(currentLogoUrl || null);
  const [uploadLogo, { isLoading }] = useUploadLogoMutation();

  const onDrop = useCallback(
    async (acceptedFiles: File[]) => {
      const file = acceptedFiles[0];
      if (!file) return;

      // Validate file size
      if (file.size > MAX_FILE_SIZE) {
        toast.error("Ukuran file maksimal 2MB");
        return;
      }

      // Create preview
      const objectUrl = URL.createObjectURL(file);
      setPreview(objectUrl);

      // Upload file
      try {
        const formData = new FormData();
        formData.append("logo", file);

        const result = await uploadLogo(formData).unwrap();
        toast.success("Logo berhasil diupload");

        // Clean up preview URL
        URL.revokeObjectURL(objectUrl);
        setPreview(result.logoUrl);
      } catch (error: any) {
        toast.error(error?.data?.error?.message || "Gagal upload logo");
        // Reset preview on error
        URL.revokeObjectURL(objectUrl);
        setPreview(currentLogoUrl || null);
      }
    },
    [uploadLogo, currentLogoUrl]
  );

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept: ACCEPTED_FILE_TYPES,
    maxFiles: 1,
    multiple: false,
    disabled: isLoading,
  });

  const handleRemove = () => {
    setPreview(null);
    // Note: Backend doesn't have delete logo endpoint yet
    toast.info("Untuk menghapus logo, upload gambar baru atau kosongkan di form");
  };

  return (
    <div className="space-y-4">
      {/* Preview */}
      {preview && (
        <div className="relative inline-block">
          <div className="relative h-32 w-48 rounded-lg border">
            <Image
              src={preview}
              alt="Company logo preview"
              fill
              className="rounded-lg object-contain"
            />
          </div>
          {!isLoading && (
            <Button
              type="button"
              variant="destructive"
              size="icon"
              className="absolute -right-2 -top-2 h-6 w-6 rounded-full"
              onClick={handleRemove}
            >
              <X className="h-3 w-3" />
            </Button>
          )}
        </div>
      )}

      {/* Upload Area */}
      <div
        {...getRootProps()}
        className={`
          relative cursor-pointer rounded-lg border-2 border-dashed p-8 text-center transition-colors
          ${isDragActive ? "border-primary bg-primary/5" : "border-muted-foreground/25"}
          ${isLoading ? "pointer-events-none opacity-50" : "hover:border-primary hover:bg-primary/5"}
        `}
      >
        <input {...getInputProps()} />

        <div className="flex flex-col items-center justify-center gap-2">
          {isLoading ? (
            <>
              <Loader2 className="h-10 w-10 animate-spin text-muted-foreground" />
              <p className="text-sm text-muted-foreground">Uploading logo...</p>
            </>
          ) : (
            <>
              {preview ? (
                <ImageIcon className="h-10 w-10 text-muted-foreground" />
              ) : (
                <Upload className="h-10 w-10 text-muted-foreground" />
              )}
              <div className="space-y-1">
                <p className="text-sm font-medium">
                  {isDragActive ? "Drop file di sini" : "Klik atau drag & drop file"}
                </p>
                <p className="text-xs text-muted-foreground">
                  JPG atau PNG, maksimal 2MB
                </p>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
