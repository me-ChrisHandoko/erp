// Tenant and User Management Types
// Aligned with backend API: /api/v1/tenant endpoints

export type UserRole = 'OWNER' | 'ADMIN' | 'STAFF' | 'VIEWER';

export type SubscriptionStatus = 'ACTIVE' | 'EXPIRED' | 'TRIAL' | 'SUSPENDED';

export interface Subscription {
  id: string;
  tenantId: string;
  planName: string;
  status: SubscriptionStatus;
  maxUsers: number;
  startDate: string;
  endDate: string;
  createdAt: string;
  updatedAt: string;
}

export interface Tenant {
  id: string;
  name: string;
  subdomain: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
  subscription?: Subscription;
}

export interface TenantUser {
  id: string;
  tenantId: string;
  email: string;
  name: string;
  phone?: string;
  role: UserRole;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
  lastLoginAt?: string;
}

export interface InviteUserRequest {
  email: string;
  name: string;
  phone?: string;
  role: UserRole;
}

export interface UpdateUserRoleRequest {
  role: UserRole;
}

export interface GetUsersFilters {
  role?: UserRole;
  isActive?: boolean;
  page?: number;
  limit?: number;
}

export interface TenantWithUsers {
  tenant: Tenant;
  users: TenantUser[];
  totalUsers: number;
  activeUsers: number;
}
