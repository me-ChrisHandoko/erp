/**
 * Company Validation Schemas
 *
 * Zod schemas for company profile and bank account validation.
 * Based on backend validation rules from internal/dto/company_dto.go
 */

import { z } from "zod";
import { ENTITY_TYPES, INDONESIAN_BANKS, INDONESIAN_PROVINCES } from "@/types/company.types";

/**
 * NPWP Validation
 * NPWP format: 15 digits, formatted as XX.XXX.XXX.X-XXX.XXX
 * Example: 12.345.678.9-012.345
 */
export const npwpSchema = z
  .string()
  .min(15, "NPWP minimal 15 karakter")
  .max(20, "NPWP maksimal 20 karakter")
  .regex(
    /^\d{2}\.\d{3}\.\d{3}\.\d-\d{3}\.\d{3}$/,
    "Format NPWP tidak valid (contoh: 12.345.678.9-012.345)"
  )
  .optional();

/**
 * Indonesian Phone Number Validation
 * Formats: +62xxx, 08xxx, 62xxx
 */
export const phoneSchema = z
  .string()
  .min(10, "Nomor telepon minimal 10 digit")
  .max(20, "Nomor telepon maksimal 20 karakter")
  .regex(
    /^(\+62|62|0)[0-9]{9,18}$/,
    "Format nomor telepon tidak valid (contoh: 08123456789 atau +628123456789)"
  );

/**
 * Email Validation
 */
export const emailSchema = z
  .string()
  .email("Format email tidak valid")
  .max(255, "Email maksimal 255 karakter");

/**
 * Website URL Validation
 */
export const websiteSchema = z
  .string()
  .url("Format URL tidak valid")
  .max(255, "URL maksimal 255 karakter")
  .optional();

/**
 * Logo URL Validation
 */
export const logoUrlSchema = z
  .string()
  .url("Format URL tidak valid")
  .max(500, "URL maksimal 500 karakter")
  .optional();

/**
 * Update Company Schema
 * All fields are optional for partial updates
 */
export const updateCompanySchema = z.object({
  name: z
    .string()
    .min(2, "Nama perusahaan minimal 2 karakter")
    .max(255, "Nama perusahaan maksimal 255 karakter")
    .optional(),

  legalName: z
    .string()
    .min(2, "Nama legal minimal 2 karakter")
    .max(255, "Nama legal maksimal 255 karakter")
    .optional(),

  entityType: z.enum(ENTITY_TYPES, {
    message: "Jenis badan usaha tidak valid"
  }).optional(),

  address: z
    .string()
    .max(500, "Alamat maksimal 500 karakter")
    .optional(),

  city: z
    .string()
    .max(100, "Nama kota maksimal 100 karakter")
    .optional(),

  province: z
    .enum(INDONESIAN_PROVINCES, {
      message: "Provinsi tidak valid"
    })
    .optional(),

  postalCode: z
    .string()
    .length(5, "Kode pos harus 5 digit")
    .regex(/^\d{5}$/, "Kode pos hanya boleh berisi angka")
    .optional(),

  phone: phoneSchema.optional(),

  email: emailSchema.optional(),

  website: websiteSchema,

  logoUrl: logoUrlSchema,

  npwp: npwpSchema,

  nib: z
    .string()
    .max(50, "NIB maksimal 50 karakter")
    .optional(),

  isPkp: z.boolean().optional(),

  ppnRate: z
    .number()
    .min(0, "PPN rate minimal 0%")
    .max(100, "PPN rate maksimal 100%")
    .optional(),

  invoicePrefix: z
    .string()
    .max(10, "Prefix invoice maksimal 10 karakter")
    .optional(),

  invoiceNumberFormat: z
    .string()
    .max(50, "Format nomor invoice maksimal 50 karakter")
    .optional(),

  fakturPajakSeries: z
    .string()
    .max(20, "Seri faktur pajak maksimal 20 karakter")
    .optional(),

  sppkpNumber: z
    .string()
    .max(50, "Nomor SPPKP maksimal 50 karakter")
    .optional(),
}).refine(
  (data) => {
    // If isPKP is true, require tax-related fields
    if (data.isPkp === true) {
      return data.npwp && data.ppnRate !== undefined;
    }
    return true;
  },
  {
    message: "Perusahaan PKP harus memiliki NPWP dan PPN Rate",
    path: ["isPkp"],
  }
);

/**
 * Add Bank Account Schema
 * All required fields for creating new bank account
 */
export const addBankAccountSchema = z.object({
  bankName: z
    .string()
    .min(2, "Nama bank minimal 2 karakter")
    .max(100, "Nama bank maksimal 100 karakter"),

  accountNumber: z
    .string()
    .min(8, "Nomor rekening minimal 8 digit")
    .max(50, "Nomor rekening maksimal 50 karakter")
    .regex(/^[0-9]+$/, "Nomor rekening hanya boleh berisi angka"),

  accountName: z
    .string()
    .min(3, "Nama pemilik rekening minimal 3 karakter")
    .max(255, "Nama pemilik rekening maksimal 255 karakter"),

  branchName: z
    .string()
    .max(255, "Nama cabang maksimal 255 karakter")
    .optional(),

  isPrimary: z.boolean().default(false),

  checkPrefix: z
    .string()
    .max(20, "Prefix cek maksimal 20 karakter")
    .optional(),
});

/**
 * Update Bank Account Schema
 * All fields optional for partial updates
 */
export const updateBankAccountSchema = z.object({
  bankName: z
    .string()
    .min(2, "Nama bank minimal 2 karakter")
    .max(100, "Nama bank maksimal 100 karakter")
    .optional(),

  accountNumber: z
    .string()
    .min(8, "Nomor rekening minimal 8 digit")
    .max(50, "Nomor rekening maksimal 50 karakter")
    .regex(/^[0-9]+$/, "Nomor rekening hanya boleh berisi angka")
    .optional(),

  accountName: z
    .string()
    .min(3, "Nama pemilik rekening minimal 3 karakter")
    .max(255, "Nama pemilik rekening maksimal 255 karakter")
    .optional(),

  branchName: z
    .string()
    .max(255, "Nama cabang maksimal 255 karakter")
    .optional(),

  isPrimary: z.boolean().optional(),

  checkPrefix: z
    .string()
    .max(20, "Prefix cek maksimal 20 karakter")
    .optional(),
});

/**
 * Bank Account Selection Schema
 * For selecting bank from dropdown
 */
export const bankSelectionSchema = z.enum(INDONESIAN_BANKS, {
  message: "Pilih bank yang valid"
});

/**
 * Type inference from schemas
 */
export type UpdateCompanyFormData = z.infer<typeof updateCompanySchema>;
export type AddBankAccountFormData = z.infer<typeof addBankAccountSchema>;
export type UpdateBankAccountFormData = z.infer<typeof updateBankAccountSchema>;
