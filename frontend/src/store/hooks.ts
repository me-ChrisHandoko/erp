// Redux Typed Hooks
// Pre-typed versions of useDispatch and useSelector for TypeScript

import { useDispatch, useSelector } from 'react-redux';
import type { RootState, AppDispatch } from './index';

/**
 * Typed version of useDispatch hook
 * Use throughout the app instead of plain `useDispatch`
 *
 * @example
 * const dispatch = useAppDispatch();
 * dispatch(login({ email, password }));
 */
export const useAppDispatch = useDispatch.withTypes<AppDispatch>();

/**
 * Typed version of useSelector hook
 * Use throughout the app instead of plain `useSelector`
 *
 * @example
 * const user = useAppSelector(state => state.auth.user);
 * const isAuthenticated = useAppSelector(state => state.auth.isAuthenticated);
 */
export const useAppSelector = useSelector.withTypes<RootState>();
