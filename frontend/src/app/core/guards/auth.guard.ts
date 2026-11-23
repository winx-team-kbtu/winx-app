import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';
import { AuthService } from '../services/auth.service';
import { safeReturnUrl } from '../services/auth-session';
import { ModeService } from '../services/mode.service';
import { OnboardingService } from '../services/onboarding.service';

export const authGuard: CanActivateFn = async (route, state) => {
  const auth = inject(AuthService);
  const router = inject(Router);
  const onboarding = inject(OnboardingService);
  auth.sessionId();
  if (!auth.isLoggedIn())
    return router.createUrlTree(['/login'], {
      queryParams: { returnUrl: safeReturnUrl(state.url) },
    });
  if (route.routeConfig?.path === 'onboarding' || route.routeConfig?.path === 'admin') return true;
  try {
    if (!(await onboarding.required())) return true;
  } catch {
    // A failed status check must not turn an existing account into a new account.
    return true;
  }
  return router.createUrlTree(['/onboarding'], {
    queryParams: { returnUrl: safeReturnUrl(state.url) },
  });
};

export const guestGuard: CanActivateFn = (route) => {
  const auth = inject(AuthService);
  const router = inject(Router);
  const mode = inject(ModeService);
  auth.sessionId();
  return (
    !auth.isLoggedIn() ||
    router.parseUrl(safeReturnUrl(route.queryParamMap.get('returnUrl')) || mode.home())
  );
};
