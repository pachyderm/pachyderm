import {useState, useRef} from 'react';

import {ModifyRoleBindingRequest, ResourceType} from '@dash-frontend/api/auth';
import {useModifyRoleBinding} from '@dash-frontend/hooks/useModifyRoleBinding';

import {UserTableRoles} from '../../../hooks/useRolesModal';
import {USER_TYPES} from '../AssignRolesForm';

const useAssignRolesForm = (
  resourceName: string,
  resourceType: ResourceType,
  availableRoles: string[],
  userTableRoles: UserTableRoles,
  deletedRoles: Record<string, ModifyRoleBindingRequest>,
  setDeletedRoles: React.Dispatch<
    React.SetStateAction<Record<string, ModifyRoleBindingRequest>>
  >,
) => {
  const inputRef = useRef<HTMLInputElement>(null);
  const [userType, setUserType] = useState(USER_TYPES[0]);
  const [email, setEmail] = useState('');
  const [role, setRole] = useState(availableRoles[0]);
  const [validationError, setValidationError] = useState('');

  const {modifyRoleBinding, loading, error} = useModifyRoleBinding({
    resource: {name: resourceName, type: resourceType},
  });

  const onSubmit = async () => {
    const user =
      userType === 'allClusterUsers'
        ? 'allClusterUsers'
        : `${userType}:${email}`;

    if (!email && userType !== 'allClusterUsers') {
      setValidationError('A name or email is required');
      return;
    }
    setValidationError('');

    modifyRoleBinding(
      {
        resource: {
          name: resourceName,
          type: resourceType,
        },
        principal: user,
        roles: [
          ...(userTableRoles[user] || {unlockedRoles: []}).unlockedRoles,
          role,
        ],
      },
      {
        onSettled: () => {
          const updatedDeletedRoles = {...deletedRoles};
          delete updatedDeletedRoles[user];
          setDeletedRoles(updatedDeletedRoles);
        },
      },
    );

    setEmail('');
    if (userType === 'allClusterUsers') {
      setUserType(USER_TYPES[0]);
    }
    if (inputRef.current) {
      inputRef.current.value = '';
    }
  };

  const hasAllClusterUsers =
    Object.keys(userTableRoles).includes('allClusterUsers');

  return {
    inputRef,
    email,
    setEmail,
    role,
    setRole,
    userType,
    setUserType,
    validationError,
    error,
    loading,
    onSubmit,
    hasAllClusterUsers,
  };
};

export default useAssignRolesForm;
