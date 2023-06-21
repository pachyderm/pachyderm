import {ResourceType, ModifyRolesArgs} from '@graphqlTypes';
import {useState, useRef} from 'react';

import {useModifyRolesMutation} from '@dash-frontend/generated/hooks';
import {GET_ROLES_QUERY} from '@dash-frontend/queries/GetRolesQuery';

import {UserTableRoles} from '../../../hooks/useRolesModal';
import {getPermissionQueries} from '../../../util/rolesUtils';
import {USER_TYPES} from '../AssignRolesForm';

const useAssignRolesForm = (
  resourceName: string,
  resourceType: ResourceType,
  availableRoles: string[],
  userTableRoles: UserTableRoles,
  deletedRoles: Record<string, ModifyRolesArgs>,
  setDeletedRoles: React.Dispatch<
    React.SetStateAction<Record<string, ModifyRolesArgs>>
  >,
) => {
  const inputRef = useRef<HTMLInputElement>(null);
  const [userType, setUserType] = useState(USER_TYPES[0]);
  const [email, setEmail] = useState('');
  const [role, setRole] = useState(availableRoles[0]);
  const [validationError, setValidationError] = useState('');

  const [modifyRolesMutation, {loading, error}] = useModifyRolesMutation({
    refetchQueries: [
      {
        query: GET_ROLES_QUERY,
        variables: {
          args: {
            resource: {name: resourceName, type: resourceType},
          },
        },
      },
      ...getPermissionQueries(resourceName, resourceType),
    ],
  });

  const onSubmit = async () => {
    if (!email) {
      setValidationError('A name or email is required');
      return;
    }
    setValidationError('');

    await modifyRolesMutation({
      variables: {
        args: {
          resource: {
            name: resourceName,
            type: resourceType,
          },
          principal: `${userType}:${email}`,
          rolesList: [
            ...(userTableRoles[`${userType}:${email}`] || {unlockedRoles: []})
              .unlockedRoles,
            role,
          ],
        },
      },
      onCompleted: () => {
        const updatedDeletedRoles = {...deletedRoles};
        delete updatedDeletedRoles[`${userType}:${email}`];
        setDeletedRoles(updatedDeletedRoles);
      },
    });

    setEmail('');
    if (inputRef.current) {
      inputRef.current.value = '';
    }
  };

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
  };
};

export default useAssignRolesForm;
