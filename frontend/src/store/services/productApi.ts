/**
 * Product API Service
 *
 * RTK Query service for product management including:
 * - Product CRUD operations
 * - Multi-unit management
 * - Supplier relationship management
 * - Product listing with filters and pagination
 *
 * Backend endpoints: /api/v1/products
 */

import { createApi } from "@reduxjs/toolkit/query/react";
import { baseQueryWithReauth } from "./authApi";
import type {
  ProductResponse,
  ProductListResponse,
  CreateProductRequest,
  UpdateProductRequest,
  ProductUnitResponse,
  CreateProductUnitRequest,
  UpdateProductUnitRequest,
  ProductSupplierResponse,
  LinkSupplierRequest,
  UpdateProductSupplierRequest,
  ProductFilters,
} from "@/types/product.types";
import type { ApiSuccessResponse } from "@/types/api";

export const productApi = createApi({
  reducerPath: "productApi",
  baseQuery: baseQueryWithReauth,
  tagTypes: ["Product", "ProductList", "ProductUnit", "ProductSupplier"],
  endpoints: (builder) => ({
    // ==================== Product CRUD ====================

    /**
     * List Products with Filters & Pagination
     * GET /api/v1/products
     */
    listProducts: builder.query<ProductListResponse, ProductFilters | void>({
      query: (filters = {}) => ({
        url: "/products",
        params: {
          search: filters.search,
          category: filters.category,
          is_active: filters.isActive,
          is_batch_tracked: filters.isBatchTracked,
          is_perishable: filters.isPerishable,
          page: filters.page || 1,
          page_size: filters.pageSize || 20,
          sort_by: filters.sortBy || "code",
          sort_order: filters.sortOrder || "asc",
        },
      }),
      providesTags: (result) =>
        result
          ? [
              ...result.data.map(({ id }) => ({
                type: "Product" as const,
                id,
              })),
              { type: "ProductList", id: "LIST" },
            ]
          : [{ type: "ProductList", id: "LIST" }],
    }),

    /**
     * Get Product Details
     * GET /api/v1/products/:id
     */
    getProduct: builder.query<ProductResponse, string>({
      query: (id) => `/products/${id}`,
      transformResponse: (response: ApiSuccessResponse<ProductResponse>) =>
        response.data,
      providesTags: (result, error, id) => [{ type: "Product", id }],
    }),

    /**
     * Create Product
     * POST /api/v1/products
     * Requires OWNER or ADMIN role
     */
    createProduct: builder.mutation<ProductResponse, CreateProductRequest>({
      query: (data) => ({
        url: "/products",
        method: "POST",
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<ProductResponse>) =>
        response.data,
      invalidatesTags: [{ type: "ProductList", id: "LIST" }],
    }),

    /**
     * Update Product
     * PUT /api/v1/products/:id
     * Requires OWNER or ADMIN role
     */
    updateProduct: builder.mutation<
      ProductResponse,
      { id: string; data: UpdateProductRequest }
    >({
      query: ({ id, data }) => ({
        url: `/products/${id}`,
        method: "PUT",
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<ProductResponse>) =>
        response.data,
      invalidatesTags: (result, error, { id }) => [
        { type: "Product", id },
        { type: "ProductList", id: "LIST" },
      ],
    }),

    /**
     * Delete Product (Soft Delete)
     * DELETE /api/v1/products/:id
     * Requires OWNER or ADMIN role
     */
    deleteProduct: builder.mutation<void, string>({
      query: (id) => ({
        url: `/products/${id}`,
        method: "DELETE",
      }),
      invalidatesTags: (result, error, id) => [
        { type: "Product", id },
        { type: "ProductList", id: "LIST" },
      ],
    }),

    // ==================== Product Units ====================

    /**
     * Add Product Unit
     * POST /api/v1/products/:id/units
     * Requires OWNER or ADMIN role
     */
    addProductUnit: builder.mutation<
      ProductUnitResponse,
      { productId: string; data: CreateProductUnitRequest }
    >({
      query: ({ productId, data }) => ({
        url: `/products/${productId}/units`,
        method: "POST",
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<ProductUnitResponse>) =>
        response.data,
      invalidatesTags: (result, error, { productId }) => [
        { type: "Product", id: productId },
        { type: "ProductUnit", id: "LIST" },
      ],
    }),

    /**
     * Update Product Unit
     * PUT /api/v1/products/:id/units/:unitId
     * Requires OWNER or ADMIN role
     */
    updateProductUnit: builder.mutation<
      ProductUnitResponse,
      { productId: string; unitId: string; data: UpdateProductUnitRequest }
    >({
      query: ({ productId, unitId, data }) => ({
        url: `/products/${productId}/units/${unitId}`,
        method: "PUT",
        body: data,
      }),
      transformResponse: (response: ApiSuccessResponse<ProductUnitResponse>) =>
        response.data,
      invalidatesTags: (result, error, { productId, unitId }) => [
        { type: "Product", id: productId },
        { type: "ProductUnit", id: unitId },
      ],
    }),

    /**
     * Delete Product Unit (Soft Delete)
     * DELETE /api/v1/products/:id/units/:unitId
     * Requires OWNER or ADMIN role
     */
    deleteProductUnit: builder.mutation<
      void,
      { productId: string; unitId: string }
    >({
      query: ({ productId, unitId }) => ({
        url: `/products/${productId}/units/${unitId}`,
        method: "DELETE",
      }),
      invalidatesTags: (result, error, { productId, unitId }) => [
        { type: "Product", id: productId },
        { type: "ProductUnit", id: unitId },
      ],
    }),

    // ==================== Product Suppliers ====================

    /**
     * Link Supplier to Product
     * POST /api/v1/products/:id/suppliers
     * Requires OWNER or ADMIN role
     */
    linkSupplier: builder.mutation<
      ProductSupplierResponse,
      { productId: string; data: LinkSupplierRequest }
    >({
      query: ({ productId, data }) => ({
        url: `/products/${productId}/suppliers`,
        method: "POST",
        body: data,
      }),
      transformResponse: (
        response: ApiSuccessResponse<ProductSupplierResponse>
      ) => response.data,
      invalidatesTags: (result, error, { productId }) => [
        { type: "Product", id: productId },
        { type: "ProductSupplier", id: "LIST" },
      ],
    }),

    /**
     * Update Product-Supplier Relationship
     * PUT /api/v1/products/:id/suppliers/:supplierId
     * Requires OWNER or ADMIN role
     */
    updateProductSupplier: builder.mutation<
      ProductSupplierResponse,
      {
        productId: string;
        supplierId: string;
        data: UpdateProductSupplierRequest;
      }
    >({
      query: ({ productId, supplierId, data }) => ({
        url: `/products/${productId}/suppliers/${supplierId}`,
        method: "PUT",
        body: data,
      }),
      transformResponse: (
        response: ApiSuccessResponse<ProductSupplierResponse>
      ) => response.data,
      invalidatesTags: (result, error, { productId, supplierId }) => [
        { type: "Product", id: productId },
        { type: "ProductSupplier", id: supplierId },
      ],
    }),

    /**
     * Remove Supplier from Product (Soft Delete)
     * DELETE /api/v1/products/:id/suppliers/:supplierId
     * Requires OWNER or ADMIN role
     */
    removeProductSupplier: builder.mutation<
      void,
      { productId: string; supplierId: string }
    >({
      query: ({ productId, supplierId }) => ({
        url: `/products/${productId}/suppliers/${supplierId}`,
        method: "DELETE",
      }),
      invalidatesTags: (result, error, { productId, supplierId }) => [
        { type: "Product", id: productId },
        { type: "ProductSupplier", id: supplierId },
      ],
    }),
  }),
});

// Export hooks for usage in components
export const {
  // Product CRUD
  useListProductsQuery,
  useGetProductQuery,
  useCreateProductMutation,
  useUpdateProductMutation,
  useDeleteProductMutation,
  // Product Units
  useAddProductUnitMutation,
  useUpdateProductUnitMutation,
  useDeleteProductUnitMutation,
  // Product Suppliers
  useLinkSupplierMutation,
  useUpdateProductSupplierMutation,
  useRemoveProductSupplierMutation,
} = productApi;
