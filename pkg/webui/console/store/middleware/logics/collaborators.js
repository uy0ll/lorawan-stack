// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import api from '@console/api'

import createRequestLogic from '@ttn-lw/lib/store/logics/create-request-logic'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import * as collaborators from '@console/store/actions/collaborators'
import { getUser } from '@console/store/actions/users'

const validParentTypes = ['application', 'gateway', 'organization']

const parentTypeValidator = ({ action }, allow) => {
  if (!validParentTypes.includes(action.payload.parentType)) {
    // Do not reject the action but throw an error, as this is an implementation
    // error.
    throw new Error(`Invalid parent entity type ${action.payload.parentType}`)
  }
  allow(action)
}

const getCollaboratorLogic = createRequestLogic({
  type: collaborators.GET_COLLABORATOR,
  validate: parentTypeValidator,
  process: async ({ action }, dispatch) => {
    const { parentType, parentId, collaboratorId, isUser } = action.payload

    if (isUser) {
      // Also retrieve the user to be able to determine whether
      // it is an admin user.
      try {
        await dispatch(attachPromise(getUser(collaboratorId, 'admin')))
      } catch {
        // Do not fail the action if the user could not be fetched.
      }
    }

    return isUser
      ? api[parentType].collaborators.getUser(parentId, collaboratorId)
      : api[parentType].collaborators.getOrganization(parentId, collaboratorId)
  },
})

const getCollaboratorsLogic = createRequestLogic({
  type: collaborators.GET_COLLABORATORS_LIST,
  validate: parentTypeValidator,
  process: async ({ action }) => {
    const { parentType, parentId, params } = action.payload
    const res = await api[parentType].collaborators.list(parentId, params)
    return { entities: res.collaborators, totalCount: res.totalCount }
  },
})

export default [getCollaboratorLogic, getCollaboratorsLogic]
